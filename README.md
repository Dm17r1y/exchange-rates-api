# Exchange service api

Api for monitoring currency exchange rates. Supports EUR, USD and MXN currencies

### Start service

#### Prerequisites

- `Go v1.25.1`
- `GNU Make v4.4.1`
- `Docker v28.2.0`
- `Docker Compose v2.39.2`

#### Start service

To start service run

```
make
```

After that, you can access your service on `http://localhost:8000`

You can show and run http methods via `http://localhost:8000/swagger/index.html` 

#### Run test

To run test, you can type

```
make test
```


#### Stop

To stop services, type

```
make stop
```
#### Cleanup

If you want to clean after run (delete images and .env.secret file), you can type

```
make clean
```