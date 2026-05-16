#pragma once
#include <string>
#include <atomic>

class EventCollector {
public:
    EventCollector(const std::string& backend_url);
    void set_agent_id(const std::string& id) { agent_id_ = id; }
    void start();
    void stop();
    bool is_running() const;

private:
    void collect_events();
    void send_event(const std::string& json_payload);

    std::string backend_url_;
    std::string agent_id_;
    std::atomic<bool> running_{false};
};