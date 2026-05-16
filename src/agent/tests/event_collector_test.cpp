#include "test_runner.h"
#include "../include/event_collector.h"

TEST(test_event_collector_initialization) {
    EventCollector ec("http://localhost:8080");
    ASSERT_EQ("", ec.get_agent_id());
    ASSERT_FALSE(ec.is_running());
    return true;
}

TEST(test_event_collector_agent_id_setter) {
    EventCollector ec("http://localhost:8080");
    ec.set_agent_id("agent-456");
    return true;
}

void register_event_collector_tests() {
    REGISTER_TEST(test_event_collector_initialization);
    REGISTER_TEST(test_event_collector_agent_id_setter);
}