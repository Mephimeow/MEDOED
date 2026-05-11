#include "event_collector.h"
#include "logger.h"
#include <curl/curl.h>
#include <thread>
#include <fstream>
#include <unordered_map>

size_t write_callback(void* contents, size_t size, size_t nmemb, void* userp) {
    return size * nmemb;
}

EventCollector::EventCollector(const std::string& backend_url)
    : backend_url_(backend_url) {}

void EventCollector::start() {
    running_ = true;
    log_info("EventCollector started");
    collect_events();
}

void EventCollector::stop() {
    running_ = false;
    log_info("EventCollector stopped");
}

bool EventCollector::is_running() const {
    return running_;
}

void EventCollector::send_event(const std::string& event) {
    CURL* curl = curl_easy_init();
    if (curl) {
        std::string url = backend_url_ + "/events";
        struct curl_slist* headers = NULL;
        headers = curl_slist_append(headers, "Content-Type: application/json");
        
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, event.c_str());
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
        
        curl_easy_perform(curl);
        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
    }
}

void EventCollector::collect_events() {
    std::ifstream proc_file("/proc/self/status");
    if (!proc_file.is_open()) {
        log_error("Cannot open /proc/self/status");
        return;
    }

    std::unordered_map<std::string, std::string> proc_stats;
    std::string line;
    while (std::getline(proc_file, line)) {
        size_t colon = line.find(':');
        if (colon != std::string::npos) {
            std::string key = line.substr(0, colon);
            std::string val = line.substr(colon + 1);
            proc_stats[key] = val;
        }
    }

    char hostname[256];
    gethostname(hostname, sizeof(hostname));
    
    std::string event = "{";
    event += "\"type\":\"process_info\",";
    event += "\"hostname\":\"" + std::string(hostname) + "\",";
    event += "\"pid\":" + proc_stats["Pid"] + ",";
    event += "\"memory\":" + proc_stats["VmRSS"].substr(1);
    event += "}";
    
    send_event(event);
    log_info("Sent process info event");
    
    while (running_) {
        std::this_thread::sleep_for(std::chrono::seconds(10));
    }
}
