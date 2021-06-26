package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4"
	"github.com/mmcloughlin/geohash"
	"github.com/uber/h3-go"
)

// Ping ... replies to a ping message for healthcheck purposes
func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

type geoEncodingRequest struct {
	Resolution uint    `json:"resolution"`
	Latitude   float64 `json:"lat"`
	Longitude  float64 `json:"long"`
	Neighbours uint    `json:"k"`
}

type geoDecodingRequest struct {
	Encoded string `json:"encoded"`
}

// EncodeGeohash returns geohash from lat long coordinates
func EncodeGeohash(c *gin.Context) {

	req := geoEncodingRequest{}
	err := c.BindJSON(&req)

	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {
		geohash := geohash.EncodeWithPrecision(req.Latitude, req.Longitude, req.Resolution)
		log.Println("encoding geohash from ", req.Latitude, req.Longitude, "to", geohash)
		c.JSON(http.StatusOK, geohash)
	}
}

// DecodeGeohash ... returns lat long coordinates from h3 string index
func DecodeGeohash(c *gin.Context) {
	req := geoDecodingRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {
		lat, long := geohash.Decode(req.Encoded)
		p := Point{lat, long}
		c.JSON(http.StatusOK, p)
	}
}

// EncodeH3 ... returns h3 from lat long coordinates
func EncodeH3(c *gin.Context) {
	req := geoEncodingRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {
		geo := h3.GeoCoord{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		}
		resolution := req.Resolution
		result := h3.FromGeo(geo, int(resolution))
		c.JSON(http.StatusOK, h3.ToString(result))
	}
}

// DecodeH3 ... returns lat long coordinates from h3 string index
func DecodeH3(c *gin.Context) {
	req := geoDecodingRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {
		index := h3.FromString(req.Encoded)
		coordinates := h3.ToGeo(index)
		c.JSON(http.StatusOK, coordinates)
	}
}

// H3Kring ... returns h3 kring from lat long coordinates
func H3Kring(c *gin.Context) {
	req := geoEncodingRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {
		geo := h3.GeoCoord{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		}
		// get all k neighbours
		neighbours := h3.KRing(h3.FromGeo(geo, int(req.Resolution)), int(req.Neighbours))
		// convert to string
		result := []string{}
		for _, e := range neighbours {
			result = append(result, h3.ToString(e))
		}

		c.JSON(http.StatusOK, result)
	}
}

type routingRequest struct {
	SrcLatitude  float64 `json:"src-lat"`
	SrcLongitude float64 `json:"src-long"`
	DstLatitude  float64 `json:"dst-lat"`
	DstLongitude float64 `json:"dst-long"`
}

type step struct {
	Seq    uint    `json:"seq"`
	Node   int64   `json:"node"`
	Street string  `json:"street"`
	Cost   float64 `json:"cost"`
	Dist_m float64 `json:"dist-m"`
}

func Dijkstra(c *gin.Context) {
	req := routingRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {

		var resultSet []*step
		query := fmt.Sprintf(`select r.seq, r.node, w.osm_name as street, r.cost::numeric(10,4), (sum(ST_Length(w.geom_way::geography)) over(order by r.seq))::numeric(10,2) as dist_m
		from pgr_dijkstra(
			'select id, source, target, cost, reverse_cost from osm_2po_4pgr',
			(select id from osm_2po_4pgr_vertices_pgr order by the_geom <-> ST_SetSRID(ST_Point(%f, %f), 4326) limit 1),
			(select id from osm_2po_4pgr_vertices_pgr order by the_geom <-> ST_SetSRID(ST_Point(%f, %f), 4326) limit 1),
			true
		) as r
		left join osm_2po_4pgr as w on r.edge = w.id;`, req.SrcLongitude, req.SrcLatitude, req.DstLongitude, req.DstLatitude)

		//model.Scan(query, resultSet)
		pgxscan.Select(context.Background(), model.connPool, &resultSet, query)

		c.JSON(http.StatusOK, resultSet)
	}
}

type point struct {
	/*
		x float64 `json:"x"`
		y float64 `json:"y"`
	*/
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type rdpRequest struct {
	Seq     []point `json:"seq"`
	Epsilon float64 `json:"epsilon"`
}

func Rdp(c *gin.Context) {
	req := &rdpRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
		return
	}

	if len(req.Seq) <= 2 {
		restErr := GetBadRequestError("need more than 2 elements in seq")
		c.JSON(restErr.Status, restErr)
		return
	}

	c.JSON(http.StatusOK, RDPSimplify(req.Seq, req.Epsilon))
}

// https://rosettacode.org/wiki/Ramer-Douglas-Peucker_line_simplification#Go
func RDPSimplify(l []point, ε float64) []point {
	x := 0
	dMax := -1.
	last := len(l) - 1
	p1 := l[0]
	p2 := l[last]
	x21 := p2.X - p1.X
	y21 := p2.Y - p1.Y
	for i, p := range l[1:last] {
		if d := math.Abs(y21*p.X - x21*p.Y + p2.X*p1.Y - p2.Y*p1.X); d > dMax {
			x = i + 1
			dMax = d
		}
	}
	if dMax > ε {
		return append(RDPSimplify(l[:x+1], ε), RDPSimplify(l[x:], ε)[1:]...)
	}
	return []point{l[0], l[len(l)-1]}
}

type demRequest struct {
	Point point  `json:"location"`
	Srid  uint16 `json:"srid"`
}

var demTable string = "public.eu_dem"

func Dem(c *gin.Context) {
	req := demRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {
		var resultSet *float64
		query := fmt.Sprintf(`SELECT ROUND(
			CAST(
			  (SELECT ST_Value(dem.rast, the_point.geom) AS height_in_meters
			  FROM %s AS dem
			  CROSS JOIN (SELECT ST_SetSRID(ST_Point(%f, %f), %d) As geom) AS the_point
			  WHERE ST_Intersects(dem.rast, ST_SetSRID(ST_Point(%f, %f), %d)))
			  AS numeric
			), 2
		  )`, demTable, req.Point.X, req.Point.Y, req.Srid, req.Point.X, req.Point.Y, req.Srid)

		//model.Scan(query, resultSet)
		pgxscan.Select(context.Background(), model.connPool, &resultSet, query)
		// todo: handle case when the point matches no area - no result available
		c.JSON(http.StatusOK, resultSet)
	}
}

type inputQuery struct {
	Sql string `json:"sql"`
}

type resultSet struct {
}

// the idea of a data service is to interact with services rather than data bases
// so that we can re-use the same data models across multiple languages and applications
// Query ... expose a data service or entity service allowing for sending generic queries to the DB
func Query(c *gin.Context) {
	query := inputQuery{}
	err := c.BindJSON(&query)

	if err != nil {
		restErr := GetBadRequestError("invalid input json format")
		c.JSON(restErr.Status, restErr)
	} else {
		rows, err := model.Query(query.Sql)
		if err != nil {
			restErr := GetBadRequestError(err.Error())
			c.JSON(restErr.Status, restErr)
		} else {
			//resultSet := model.Iterate(rows, targetType)
			result := []resultSet{}
			for (*rows).Next() {
				rs := resultSet{}
				err := (*rows).Scan(&rs)
				if err != nil {
					restErr := GetBadRequestError(err.Error())
					c.JSON(restErr.Status, restErr)
					return
				}
				result = append(result, rs)
			}
			c.JSON(http.StatusOK, result)
		}
	}
}

var model *dao

func StartEndpoint() {
	model = NewDao()
	err := model.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	log.Println("Connected to database")

	var router = gin.Default()

	router.GET("healthcheck/", Ping)

	router.POST("query/", Query)

	router.POST("encodegeohash/", EncodeGeohash)
	router.POST("decodegeohash/", DecodeGeohash)

	router.POST("encodeh3/", EncodeH3)
	router.POST("h3kring/", H3Kring)
	router.POST("decodeh3/", DecodeH3)

	router.POST("dijkstra", Dijkstra)
	router.POST("rdp", Rdp)
	router.POST("dem", Dem)

	// run router as standalone service
	router.Run(fmt.Sprintf(":%s", port))
}

var port = "8085"

//os.Getenv("PORT")

func CloseEndpoint() {
	model.Close()
}
