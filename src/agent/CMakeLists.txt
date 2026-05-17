cmake_minimum_required(VERSION 3.10)
project(EDRAgent)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include_directories(include)
include_directories(tests)
include_directories(/usr/include/x86_64-linux-gnu)

add_executable(agent
    src/main.cpp
    src/logger.cpp
    src/heartbeat.cpp
    src/event_collector.cpp
    src/config.cpp
)

target_link_libraries(agent curl pthread)

enable_testing()

add_executable(agent_tests
    tests/main_test.cpp
    tests/logger_test.cpp
    tests/heartbeat_test.cpp
    tests/event_collector_test.cpp
    tests/json_test.cpp
    tests/system_test.cpp
    tests/config_test.cpp
    src/logger.cpp
    src/heartbeat.cpp
    src/event_collector.cpp
    src/config.cpp
)

target_link_libraries(agent_tests curl pthread)

add_test(NAME agent_tests COMMAND agent_tests)