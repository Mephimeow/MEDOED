#pragma once
#include <string>

class EventCollector {
public:
    EventCollector(const std::string& backend_url);
    void start();
    void stop();
    bool is_running() const;

private:
    void collect_events();
    void send_event(const std::string& event);

    std::string backend_url_;
    bool running_ = false;
};
