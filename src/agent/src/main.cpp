#include <iostream>
#include <csignal>
#include <thread>
#include <chrono>
#include <atomic>
#include "logger.h"
#include "heartbeat.h"
#include "event_collector.h"

std::atomic<bool> g_running(true);

void signal_handler(int signal) {
    if (signal == SIGINT || signal == SIGTERM) {
        g_running = false;
        log_info("Received shutdown signal");
    }
}

std::string get_backend_url() {
    const char* env = std::getenv("BACKEND_URL");
    if (env) {
        return std::string(env);
    }
    return "http://127.0.0.1:8080";
}

int main() {
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);

    log_info("MEDOED EDR Agent starting...");
    
    std::string backend_url = get_backend_url();
    log_info("Backend URL: " + backend_url);

    Heartbeater heartbeater(backend_url);
    EventCollector collector(backend_url);

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