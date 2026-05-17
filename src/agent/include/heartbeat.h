#pragma once
#include <string>
#include <atomic>

class Heartbeater {
public:
    Heartbeater(const std::string& backend_url);
    void start();
    void stop();
    bool is_running() const;
    std::string get_agent_id() const { return agent_id_; }
    void set_agent_id(const std::string& id) { agent_id_ = id; }
    void set_heartbeat_interval(int seconds) { heartbeat_interval_ = seconds; }

private:
    void register_agent();
    void send_heartbeat();

    std::string backend_url_;
    std::string agent_id_;
    std::atomic<bool> running_{false};
    std::atomic<bool> registered_{false};
    int heartbeat_interval_ = 5;
};