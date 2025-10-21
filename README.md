# Exchange service api

Api for monitoring currency exchange rates. Supports EUR, USD, and MXN currencies

### Start service

#### Prerequisites

- `Go v1.25.1`
- `GNU Make v4.3`
- `Docker v28.2.0`
- `Docker Compose v2.33.1`

#### Start service

To start the service, run

```
make
```

After that, you can access your service on `http://localhost:8080`

You can show and run HTTP methods via `http://localhost:8080/swagger/index.html` 

#### Run test

To run tests, you can type

```
make test
```


#### Stop

To stop services, type

```
make stop
```
#### Cleanup

If you want to clean after running (delete images and .env.secret files), you can type

```
make clean
```