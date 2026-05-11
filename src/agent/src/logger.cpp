#include "logger.h"
#include <iostream>
#include <chrono>
#include <iomanip>
#include <sstream>

std::string current_timestamp() {
    auto now = std::chrono::system_clock::now();
    auto time = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
        now.time_since_epoch()) % 1000;
    
    std::ostringstream oss;
    oss << std::put_time(std::localtime(&time), "%Y-%m-%d %H:%M:%S");
    oss << '.' << std::setfill('0') << std::setw(3) << ms.count();
    return oss.str();
}

void log_info(const std::string& msg) {
    std::cout << "[" << current_timestamp() << "] [INFO] " << msg << std::endl;
}

void log_error(const std::string& msg) {
    std::cerr << "[" << current_timestamp() << "] [ERROR] " << msg << std::endl;
}

void log_debug(const std::string& msg) {
    std::cout << "[" << current_timestamp() << "] [DEBUG] " << msg << std::endl;
}
