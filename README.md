# Wasmer Leak

This is a demo example of wasmer-go to reproduce a possible memory leak.

1. First build the container using `docker-compose build`
2. Start e.g. `docker stats` to track the memory usage
3. Start the container using `docker-compose --compatibility up`
4. The memory will quickly ramp up, notice that we set a memory limit of 200MB in the docker-compose file. It runs a single wasm function in an eternal loop.
