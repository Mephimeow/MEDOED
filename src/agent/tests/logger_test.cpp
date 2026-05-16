#include "test_runner.h"
#include "../include/logger.h"
#include <fstream>
#include <sstream>

TEST(test_logger_info) {
    log_info("Test info message");
    return true;
}

TEST(test_logger_error) {
    log_error("Test error message");
    return true;
}

TEST(test_logger_debug) {
    log_debug("Test debug message");
    return true;
}

void register_logger_tests() {
    REGISTER_TEST(test_logger_info);
    REGISTER_TEST(test_logger_error);
    REGISTER_TEST(test_logger_debug);
}