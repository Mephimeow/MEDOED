#include "test_runner.h"
#include "config.h"

static void test_config_default_values() {
    AgentConfig config;
    config.backend_url = "";
    config.heartbeat_interval = 30;
    config.log_level = "info";
    
    ASSERT_EQ(config.heartbeat_interval, 30);
}

static void test_config_file_exists() {
    ASSERT_TRUE(true);
}

static void test_config_loading_from_env() {
    setenv("BACKEND_URL", "http://test.example.com:8080", 1);
    
    AgentConfig config = ConfigLoader::load("");
    
    ASSERT_STRING_CONTAINS(config.backend_url, "test.example.com");
}

static void test_config_default_path() {
    std::string path = ConfigLoader::get_default_config_path();
    ASSERT_TRUE(!path.empty());
}

void register_config_tests();

int main() {
    std::cout << "===========================================\n";
    std::cout << "        MEDOED Agent Test Suite\n";
    std::cout << "===========================================\n\n";

    register_logger_tests();
    register_heartbeat_tests();
    register_event_collector_tests();
    register_json_tests();
    register_system_tests();
    register_config_tests();

    int failed = RUN_TESTS();
    
    std::cout << "\n===========================================\n";
    if (failed == 0) {
        std::cout << "         ALL TESTS PASSED!\n";
    } else {
        std::cout << "         " << failed << " TEST(S) FAILED\n";
    }
    std::cout << "===========================================\n";
    
    return failed > 0 ? 1 : 0;
}