#include "event_collector.h"
#include "heartbeat.h"
#include "logger.h"
#include <curl/curl.h>
#include <thread>
#include <fstream>
#include <sstream>
#include <iomanip>
#include <ctime>
#include <unistd.h>
#include <dirent.h>
#include <sys/stat.h>

static size_t write_callback(void* contents, size_t size, size_t nmemb, std::string* output) {
    size_t total_size = size * nmemb;
    output->append((char*)contents, total_size);
    return total_size;
}

std::string get_timestamp() {
    auto now = std::chrono::system_clock::now();
    auto time = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
        now.time_since_epoch()) % 1000;
    
    std::ostringstream oss;
    oss << std::put_time(std::localtime(&time), "%Y-%m-%dT%H:%M:%S");
    oss << '.' << std::setfill('0') << std::setw(3) << ms.count() << "Z";
    return oss.str();
}

std::string escape_json(const std::string& s) {
    std::ostringstream oss;
    for (char c : s) {
        switch (c) {
            case '"': oss << "\\\""; break;
            case '\\': oss << "\\\\"; break;
            case '\b': oss << "\\b"; break;
            case '\f': oss << "\\f"; break;
            case '\n': oss << "\\n"; break;
            case '\r': oss << "\\r"; break;
            case '\t': oss << "\\t"; break;
            default: oss << c;
        }
    }
    return oss.str();
}

EventCollector::EventCollector(const std::string& backend_url)
    : backend_url_(backend_url) {}

void EventCollector::start() {
    running_ = true;
    log_info("EventCollector started");
    
    std::string hostname;
    char buf[256];
    gethostname(buf, sizeof(buf));
    hostname = buf;
    
    collect_process_events(hostname);
    
    while (running_) {
        std::this_thread::sleep_for(std::chrono::seconds(30));
        if (running_ && !agent_id_.empty()) {
            collect_process_events(hostname);
        }
    }
}

void EventCollector::stop() {
    running_ = false;
    log_info("EventCollector stopped");
}

bool EventCollector::is_running() const {
    return running_;
}

void EventCollector::send_event(const std::string& json_payload) {
    if (agent_id_.empty()) {
        log_error("Cannot send event: agent not registered");
        return;
    }
    
    CURL* curl = curl_easy_init();
    if (!curl) return;
    
    std::string url = backend_url_ + "/api/v1/events";
    
    struct curl_slist* headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    
    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_payload.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    
    CURLcode res = curl_easy_perform(curl);
    if (res != CURLE_OK) {
        log_error("Failed to send event: " + std::string(curl_easy_strerror(res)));
    }
    
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);
}

void EventCollector::collect_process_events(const std::string& hostname) {
    DIR* proc_dir = opendir("/proc");
    if (!proc_dir) {
        log_error("Cannot open /proc directory");
        return;
    }
    
    struct dirent* entry;
    int process_count = 0;
    int suspicious_count = 0;
    
    while ((entry = readdir(proc_dir)) != nullptr) {
        if (entry->d_name[0] < '0' || entry->d_name[0] > '9') {
            continue;
        }
        
        int pid = atoi(entry->d_name);
        if (pid <= 1) continue;
        
        std::string stat_path = "/proc/" + std::string(entry->d_name) + "/stat";
        std::ifstream stat_file(stat_path);
        
        if (!stat_file.is_open()) continue;
        
        std::string stat_content;
        std::getline(stat_file, stat_content);
        
        size_t first_paren = stat_content.find('(');
        size_t last_paren = stat_content.rfind(')');
        
        if (first_paren != std::string::npos && last_paren != std::string::npos) {
            std::string comm = stat_content.substr(first_paren + 1, last_paren - first_paren - 1);
            std::istringstream iss(stat_content.substr(last_paren + 2));
            
            std::string state;
            long utime = 0, stime = 0;
            iss >> state >> utime >> stime;
            
            process_count++;
            
            const std::string suspicious_procs[] = {
                "nc", "netcat", "ncat", "socat",
                "python", "perl", "ruby", "php",
                "bash", "sh", "zsh", "fish",
                "wget", "curl", "ncat"
            };
            
            for (const auto& sus : suspicious_procs) {
                if (comm.find(sus) != std::string::npos && 
                    (comm.find("python") != std::string::npos || 
                     comm.find("bash") != std::string::npos ||
                     comm.find("nc") != std::string::npos)) {
                    suspicious_count++;
                    
                    std::string json = "{";
                    json += "\"agent_id\":\"" + agent_id_ + "\",";
                    json += "\"event_type\":\"suspicious_process\",";
                    json += "\"timestamp\":\"" + get_timestamp() + "\",";
                    json += "\"severity\":\"warning\",";
                    json += "\"source\":\"process_monitor\",";
                    json += "\"description\":\"Suspicious process detected\",";
                    json += "\"payload\":{";
                    json += "\"pid\":" + std::to_string(pid) + ",";
                    json += "\"comm\":\"" + escape_json(comm) + "\",";
                    json += "\"state\":\"" + state + "\",";
                    json += "\"hostname\":\"" + escape_json(hostname) + "\"";
                    json += "}";
                    json += "}";
                    
                    send_event(json);
                    break;
                }
            }
        }
    }
    
    closedir(proc_dir);
    
    std::string summary_json = "{";
    summary_json += "\"agent_id\":\"" + agent_id_ + "\",";
    summary_json += "\"event_type\":\"process_snapshot\",";
    summary_json += "\"timestamp\":\"" + get_timestamp() + "\",";
    summary_json += "\"severity\":\"info\",";
    summary_json += "\"source\":\"process_monitor\",";
    summary_json += "\"description\":\"Process scan completed\",";
    summary_json += "\"payload\":{";
    summary_json += "\"total_processes\":" + std::to_string(process_count) + ",";
    summary_json += "\"suspicious_count\":" + std::to_string(suspicious_count) + ",";
    summary_json += "\"hostname\":\"" + escape_json(hostname) + "\"";
    summary_json += "}";
    summary_json += "}";
    
    send_event(summary_json);
    log_info("Process scan completed: " + std::to_string(process_count) + " processes, " + 
             std::to_string(suspicious_count) + " suspicious");
}