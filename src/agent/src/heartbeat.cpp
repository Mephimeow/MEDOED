#include "heartbeat.h"
#include "logger.h"
#include <curl/curl.h>
#include <thread>
#include <cstring>
#include <fstream>
#include <sstream>
#include <unistd.h>
#include <sys/utsname.h>

static size_t write_callback(void* contents, size_t size, size_t nmemb, std::string* output) {
    size_t total_size = size * nmemb;
    output->append((char*)contents, total_size);
    return total_size;
}

std::string get_hostname() {
    char hostname[256];
    gethostname(hostname, sizeof(hostname));
    return std::string(hostname);
}

std::string get_os_info() {
    std::ifstream os_release("/etc/os-release");
    std::string line;
    while (std::getline(os_release, line)) {
        if (line.find("PRETTY_NAME=") == 0) {
            size_t pos = line.find('=');
            if (pos != std::string::npos) {
                return line.substr(pos + 1);
            }
        }
    }
    return "Linux";
}

std::string get_kernel_version() {
    struct utsname uts;
    if (uname(&uts) == 0) {
        return std::string(uts.release);
    }
    return "unknown";
}

std::string get_ip_address() {
    std::ifstream ifaddr("/proc/net/tcp");
    std::string line;
    while (std::getline(ifaddr, line)) {
        std::istringstream iss(line);
        std::string local_address;
        iss >> local_address;
        if (local_address.find("0100007F") != std::string::npos) {
            continue;
        }
        if (local_address.length() == 8) {
            unsigned int ip = std::stoul(local_address, nullptr, 16);
            char buf[32];
            snprintf(buf, sizeof(buf), "%u.%u.%u.%u",
                (ip >> 24) & 0xFF, (ip >> 16) & 0xFF, (ip >> 8) & 0xFF, ip & 0xFF);
            return std::string(buf);
        }
    }
    return "127.0.0.1";
}

Heartbeater::Heartbeater(const std::string& backend_url)
    : backend_url_(backend_url) {}

void Heartbeater::start() {
    running_ = true;
    log_info("Heartbeater started");
    
    register_agent();
    
    while (running_) {
        send_heartbeat();
        
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

void Heartbeater::register_agent() {
    CURL* curl = curl_easy_init();
    if (!curl) return;
    
    std::string json = "{";
    json += "\"hostname\":\"" + get_hostname() + "\",";
    json += "\"os_info\":\"" + get_os_info() + "\",";
    json += "\"kernel_version\":\"" + get_kernel_version() + "\",";
    json += "\"ip_address\":\"" + get_ip_address() + "\"";
    json += "}";
    
    std::string url = backend_url_ + "/api/v1/agents/register";
    std::string response;
    
    struct curl_slist* headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    
    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
    
    CURLcode res = curl_easy_perform(curl);
    if (res == CURLE_OK) {
        size_t id_start = response.find("\"id\":\"");
        if (id_start != std::string::npos) {
            id_start += 6;
            size_t id_end = response.find("\"", id_start);
            if (id_end != std::string::npos) {
                agent_id_ = response.substr(id_start, id_end - id_start);
                registered_ = true;
                log_info("Agent registered with ID: " + agent_id_);
            }
        }
    } else {
        log_error("Agent registration failed: " + std::string(curl_easy_strerror(res)));
    }
    
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);
}

void Heartbeater::send_heartbeat() {
    if (!registered_) {
        register_agent();
        return;
    }
    
    CURL* curl = curl_easy_init();
    if (!curl) return;
    
    std::string json = "{";
    json += "\"hostname\":\"" + get_hostname() + "\",";
    json += "\"os_info\":\"" + get_os_info() + "\",";
    json += "\"kernel_version\":\"" + get_kernel_version() + "\",";
    json += "\"ip_address\":\"" + get_ip_address() + "\"";
    json += "}";
    
    std::string url = backend_url_ + "/api/v1/agents/heartbeat";
    
    struct curl_slist* headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    
    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    
    CURLcode res = curl_easy_perform(curl);
    if (res != CURLE_OK) {
        log_error("Heartbeat failed: " + std::string(curl_easy_strerror(res)));
        registered_ = false;
    }
    
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);
}