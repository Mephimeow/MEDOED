#include "test_runner.h"
#include <iostream>

extern void register_logger_tests();
extern void register_heartbeat_tests();
extern void register_event_collector_tests();
extern void register_json_tests();
extern void register_system_tests();

int main() {
    std::cout << "===========================================\n";
    std::cout << "        MEDOED Agent Test Suite\n";
    std::cout << "===========================================\n\n";

    register_logger_tests();
    register_heartbeat_tests();
    register_event_collector_tests();
    register_json_tests();
    register_system_tests();

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