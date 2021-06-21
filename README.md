# Nóatún

Spatial Go(lang) service wrapping Postgis/Pg-routing data bases.

## Build

```
GO111MODULE=on go build -o noatun . 
```

## Usage

```
export DATABASE_URL="postgres://user:secret@host:port/db"
./noatun
```

## Examples

### Geohash encoding

POST to `host:8085/encodegeohash`
```json
{
    "resolution" : 7,
    "lat" : 48.6687, 
    "long" :  -4.3293
}
```

returns
```
gbsuv7z
```


### H3 encoding

POST to `host:8085/encodeh3`
```json
{
    "resolution" : 7,
    "lat" : 48.6687, 
    "long" :  -4.3293
}
```

returns
```
871871725ffffff
```

### Dijkstra

POST to `host:8085/dijkstra` with:
```json
{
    "src-lat" : 64.131259,
    "src-long" : -21.896194,
    "dst-lat": 64.128251,
    "dst-long":-21.894693
}
```

Returned response:
```json
[
    {
        "seq": 1,
        "node": 15535,
        "street": "Kringlan",
        "cost": 0.0021,
        "dist-m": 105.96
    },
    {
        "seq": 2,
        "node": 14181,
        "street": "Kringlan",
        "cost": 0.0006,
        "dist-m": 136.71
    },
    {
        "seq": 3,
        "node": 14215,
        "street": "Kringlan",
        "cost": 0.001,
        "dist-m": 187.09
    },
    {
        "seq": 4,
        "node": 14177,
        "street": "Kringlan",
        "cost": 0.0004,
        "dist-m": 208.99
    },
    {
        "seq": 5,
        "node": 14196,
        "street": "Kringlan",
        "cost": 0.0003,
        "dist-m": 225.85
    },
    {
        "seq": 6,
        "node": 8198,
        "street": "Kringlan",
        "cost": 0.0006,
        "dist-m": 257.12
    },
    {
        "seq": 7,
        "node": 15072,
        "street": "Kringlan",
        "cost": 0.0006,
        "dist-m": 285.48
    },
    {
        "seq": 8,
        "node": 44087,
        "street": "Kringlan",
        "cost": 0.0001,
        "dist-m": 290.65
    },
    {
        "seq": 9,
        "node": 947,
        "street": "Kringlan",
        "cost": 0.0003,
        "dist-m": 307.1
    },
    {
        "seq": 10,
        "node": 44092,
        "street": "Listabraut",
        "cost": 0.0011,
        "dist-m": 361.32
    },
    {
        "seq": 11,
        "node": 59453,
        "street": "Listabraut",
        "cost": 0.0011,
        "dist-m": 415.17
    },
    {
        "seq": 12,
        "node": 5794,
        "street": "Listabraut",
        "cost": 0.0006,
        "dist-m": 447.32
    },
    {
        "seq": 13,
        "node": 19372,
        "street": "Listabraut",
        "cost": 0.0005,
        "dist-m": 470.7
    },
    {
        "seq": 14,
        "node": 44095,
        "street": "Listabraut",
        "cost": 0.0001,
        "dist-m": 476.48
    },
    {
        "seq": 15,
        "node": 16552,
        "street": "Listabraut",
        "cost": 0.0001,
        "dist-m": 480.31
    },
    {
        "seq": 16,
        "node": 16549,
        "street": "Kringlan",
        "cost": 0.0002,
        "dist-m": 488.81
    },
    {
        "seq": 17,
        "node": 44096,
        "street": "Listabraut",
        "cost": 0.0002,
        "dist-m": 498.14
    }
]
```

### RDP

POST to `host:8085/rdp` (see [here](https://rosettacode.org/wiki/Ramer-Douglas-Peucker_line_simplification#Go)) with:
```json
{
    "epsilon" : 1,
    "seq" : [
        {"x":0, "y": 0}, {"x":1, "y":0.1}, {"x":2, "y":-0.1}, {"x":3, "y":5}, {"x":4, "y":6}, {"x":5, "y":7}, {"x":6, "y":8.1}, {"x":7, "y":9}, {"x":8, "y":9}, {"x":9, "y":9}
    ]
}
```

Returned response:
```json
[
    {
        "x": 0,
        "y": 0
    },
    {
        "x": 2,
        "y": -0.1
    },
    {
        "x": 3,
        "y": 5
    },
    {
        "x": 7,
        "y": 9
    },
    {
        "x": 9,
        "y": 9
    }
]
```