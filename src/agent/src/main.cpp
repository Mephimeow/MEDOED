#include <iostream>
#include <csignal>
#include <thread>
#include <chrono>
#include <atomic>
#include "logger.h"
#include "config.h"
#include "heartbeat.h"
#include "event_collector.h"

std::atomic<bool> g_running(true);

void signal_handler(int signal) {
    if (signal == SIGINT || signal == SIGTERM) {
        g_running = false;
        log_info("Received shutdown signal");
    }
}

int main(int argc, char* argv[]) {
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);

    log_info("MEDOED EDR Agent starting...");
    
    std::string config_path;
    if (argc > 1) {
        config_path = argv[1];
    }
    
    AgentConfig config = ConfigLoader::load(config_path);
    log_info("Backend URL: " + config.backend_url);

    Heartbeater heartbeater(config.backend_url);
    heartbeater.set_heartbeat_interval(config.heartbeat_interval);
    EventCollector collector(config.backend_url);

    std::thread heartbeat_thread([&heartbeater, &collector]() {
        heartbeater.start();
        collector.set_agent_id(heartbeater.get_agent_id());
    });

    std::thread collector_thread([&collector, &heartbeater]() {
        while (heartbeater.get_agent_id().empty() && g_running) {
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
        }
        collector.set_agent_id(heartbeater.get_agent_id());
        collector.start();
    });

    while (g_running) {
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }

    log_info("Shutting down...");
    heartbeater.stop();
    collector.stop();

    heartbeat_thread.join();
    collector_thread.join();

    log_info("Agent stopped");
    return 0;
}