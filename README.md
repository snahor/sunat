RUC
===

[![Build Status](https://drone.io/github.com/snahor/ruc/status.png)](https://drone.io/github.com/snahor/ruc/latest)

Overview
--------
RUC is a wrapper around SUNAT's ["Consulta de RUC"][1]. 

**It's still in development. Not fully tested.**


How to install
--------------
Ingredients:

- `Go 1+`
- `Tesseract 3.03+`

To install this server:

```
go get github.com/snahor/ruc
go build
go install
```
An executable file called `ruc` will be generated. To run it:
```
ruc 
```
By default it will listen on `127.0.0.1:8888`. You can specify host and port like:
```
ruc --host=0.0.0.0 --port=6666
```

Reference
---------
The server has two endpoints: `/search` and `/detail`.

`/search` needs a `q` parameter via query string. It will search only for RUC, DNI and names.

Example:
```
/search?q=foo%20bar
{
    "_meta": {
        "page": 0,
        "per_page": 0,
        "total": 0
    },
    "items": [
        {
            "name": "JOHN DOE",
            "location": "LIMA",
            "ruc": "12345678901",
            "status": "ACTIVO"
        }
    ]
}
```

`/detail/{{ RUC }}` needs a well formed RUC number (11 digits). If it's not valid (for instance not enough digits), a status 400 will be returned with an error message in the body. If it doesn't exist the server will return a status 404 with an error message in the body.

Example:
```
/detail/12345678901

{
    "address": "Forgiven Av. 666",
    "condition": "HABIDO",
    "dni": "12345678",
    "name": "JOHN DOE",
    "ruc": "12345678901",
    "status": "ACTIVO",
    "type": "PERSONA NATURAL SIN NEGOCIO"
}
```

TODO
----
- Pagination!
- Use channels.

  [1]: sunat.gob.pe/cl-ti-itmrconsruc/jcrS00Alias
