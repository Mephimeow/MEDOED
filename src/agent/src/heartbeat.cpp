#include "heartbeat.h"
#include "logger.h"
#include <curl/curl.h>
#include <thread>
#include <atomic>

size_t write_callback(void* contents, size_t size, size_t nmemb, void* userp) {
    return size * nmemb;
}

Heartbeater::Heartbeater(const std::string& backend_url) 
    : backend_url_(backend_url) {}

void Heartbeater::start() {
    running_ = true;
    log_info("Heartbeater started");
    
    while (running_) {
        CURL* curl = curl_easy_init();
        if (curl) {
            std::string url = backend_url_ + "/heartbeat";
            curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
            curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
            
            CURLcode res = curl_easy_perform(curl);
            if (res != CURLE_OK) {
                log_error("Heartbeat failed: " + std::string(curl_easy_strerror(res)));
            }
            curl_easy_cleanup(curl);
        }
        
        for (int i = 0; i < heartbeat_interval_ && running_; ++i) {
            std::this_thread::sleep_for(std::chrono::seconds(1));
        }
    }
}

void Heartbeater::stop() {
    running_ = false;
    log_info("Heartbeater stopped");
}

bool Heartbeater::is_running() const {
    return running_;
}
