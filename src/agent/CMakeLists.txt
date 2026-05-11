cmake_minimum_required(VERSION 3.10)
project(EDRAgent)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include_directories(include)

add_executable(agent
    src/main.cpp
    src/logger.cpp
    src/heartbeat.cpp
    src/event_collector.cpp
)

target_link_libraries(agent curl pthread)
