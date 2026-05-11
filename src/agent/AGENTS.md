# MEDOED EDR Agent

## Build
```bash
mkdir build && cd build && cmake .. && make
```

## Lint
```bash
cppcheck --enable=all src/
```

## Run Tests
```bash
./build/agent
```
