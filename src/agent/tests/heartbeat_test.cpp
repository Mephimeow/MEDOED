#include "test_runner.h"
#include "../include/heartbeat.h"
#include <cstring>

TEST(test_heartbeat_initialization) {
    Heartbeater hb("http://localhost:8080");
    ASSERT_EQ("", hb.get_agent_id());
    ASSERT_FALSE(hb.is_running());
    return true;
}

TEST(test_heartbeat_agent_id_setter) {
    Heartbeater hb("http://localhost:8080");
    hb.set_agent_id("test-agent-123");
    ASSERT_EQ("test-agent-123", hb.get_agent_id());
    return true;
}

TEST(test_heartbeat_url) {
    Heartbeater hb("http://example.com:9090");
    hb.set_agent_id("test");
    ASSERT_EQ("test", hb.get_agent_id());
    return true;
}

void register_heartbeat_tests() {
    REGISTER_TEST(test_heartbeat_initialization);
    REGISTER_TEST(test_heartbeat_agent_id_setter);
    REGISTER_TEST(test_heartbeat_url);
}