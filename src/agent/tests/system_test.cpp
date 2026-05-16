#include "test_runner.h"
#include <fstream>
#include <sstream>
#include <unistd.h>
#include <sys/utsname.h>
#include <dirent.h>
#include <sys/stat.h>

std::string get_hostname_test() {
    char hostname[256];
    gethostname(hostname, sizeof(hostname));
    return std::string(hostname);
}

std::string get_os_info_test() {
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

std::string get_kernel_version_test() {
    struct utsname uts;
    if (uname(&uts) == 0) {
        return std::string(uts.release);
    }
    return "unknown";
}

TEST(test_get_hostname) {
    std::string hostname = get_hostname_test();
    ASSERT_TRUE(hostname.length() > 0);
    ASSERT_TRUE(hostname.length() < 256);
    return true;
}

TEST(test_get_os_info) {
    std::string os_info = get_os_info_test();
    ASSERT_TRUE(os_info.length() > 0);
    return true;
}

TEST(test_get_kernel_version) {
    std::string kernel = get_kernel_version_test();
    ASSERT_TRUE(kernel.length() > 0);
    ASSERT_TRUE(kernel != "unknown");
    ASSERT_TRUE(kernel.find(".") != std::string::npos);
    return true;
}

TEST(test_proc_filesystem_accessible) {
    std::ifstream status_file("/proc/self/status");
    ASSERT_TRUE(status_file.is_open());
    
    std::string line;
    bool found_pid = false;
    while (std::getline(status_file, line)) {
        if (line.find("Pid:") == 0) {
            found_pid = true;
            break;
        }
    }
    ASSERT_TRUE(found_pid);
    return true;
}

TEST(test_read_proc_stat) {
    std::ifstream stat_file("/proc/self/stat");
    ASSERT_TRUE(stat_file.is_open());
    
    std::string content;
    std::getline(stat_file, content);
    ASSERT_TRUE(content.length() > 0);
    ASSERT_TRUE(content.find("(") != std::string::npos);
    ASSERT_TRUE(content.find(")") != std::string::npos);
    return true;
}

TEST(test_proc_directory_accessible) {
    DIR* dir = opendir("/proc");
    ASSERT_TRUE(dir != nullptr);
    closedir(dir);
    return true;
}

TEST(test_process_count) {
    DIR* proc_dir = opendir("/proc");
    ASSERT_TRUE(proc_dir != nullptr);
    
    int count = 0;
    struct dirent* entry;
    while ((entry = readdir(proc_dir)) != nullptr) {
        if (entry->d_name[0] >= '0' && entry->d_name[0] <= '9') {
            count++;
        }
    }
    closedir(proc_dir);
    
    ASSERT_TRUE(count > 0);
    return true;
}

void register_system_tests() {
    REGISTER_TEST(test_get_hostname);
    REGISTER_TEST(test_get_os_info);
    REGISTER_TEST(test_get_kernel_version);
    REGISTER_TEST(test_proc_filesystem_accessible);
    REGISTER_TEST(test_read_proc_stat);
    REGISTER_TEST(test_proc_directory_accessible);
    REGISTER_TEST(test_process_count);
}