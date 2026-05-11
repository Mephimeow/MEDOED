#pragma once
#include <string>

class Heartbeater {
public:
    Heartbeater(const std::string& backend_url);
    void start();
    void stop();
    bool is_running() const;

private:
    std::string backend_url_;
    bool running_ = false;
    int heartbeat_interval_ = 5;
};
