#include "test_runner.h"
#include <vector>
#include <sstream>
#include <iomanip>
#include <ctime>
#include <chrono>

std::string escape_json_string(const std::string& s) {
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

TEST(test_json_escape_quote) {
    std::string input = "hello \"world\"";
    std::string expected = "hello \\\"world\\\"";
    ASSERT_EQ(expected, escape_json_string(input));
    return true;
}

TEST(test_json_escape_backslash) {
    std::string input = "path\\to\\file";
    std::string expected = "path\\\\to\\\\file";
    ASSERT_EQ(expected, escape_json_string(input));
    return true;
}

TEST(test_json_escape_newline) {
    std::string input = "line1\nline2";
    std::string expected = "line1\\nline2";
    ASSERT_EQ(expected, escape_json_string(input));
    return true;
}

TEST(test_json_escape_tab) {
    std::string input = "col1\tcol2";
    std::string expected = "col1\\tcol2";
    ASSERT_EQ(expected, escape_json_string(input));
    return true;
}

TEST(test_json_escape_mixed) {
    std::string input = "quote: \" and backslash: \\ and newline:\n";
    std::string expected = "quote: \\\" and backslash: \\\\ and newline:\\n";
    ASSERT_EQ(expected, escape_json_string(input));
    return true;
}

TEST(test_json_no_escape_needed) {
    std::string input = "hello world 123";
    ASSERT_EQ(input, escape_json_string(input));
    return true;
}

std::string get_test_timestamp() {
    auto now = std::chrono::system_clock::now();
    auto time = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
        now.time_since_epoch()) % 1000;
    
    std::ostringstream oss;
    oss << std::put_time(std::localtime(&time), "%Y-%m-%dT%H:%M:%S");
    oss << '.' << std::setfill('0') << std::setw(3) << ms.count() << "Z";
    return oss.str();
}

TEST(test_timestamp_format) {
    std::string ts = get_test_timestamp();
    
    ASSERT_TRUE(ts.length() > 0);
    ASSERT_TRUE(ts[0] == '2');
    ASSERT_TRUE(ts[4] == '-');
    ASSERT_TRUE(ts[7] == '-');
    ASSERT_TRUE(ts[10] == 'T');
    ASSERT_TRUE(ts[13] == ':');
    ASSERT_TRUE(ts[16] == ':');
    ASSERT_TRUE(ts[19] == '.');
    ASSERT_TRUE(ts[ts.length() - 1] == 'Z');
    
    return true;
}

TEST(test_build_json_object) {
    std::string json = "{";
    json += "\"key1\":\"value1\",";
    json += "\"key2\":123,";
    json += "\"key3\":true";
    json += "}";
    
    ASSERT_TRUE(json.find("\"key1\":\"value1\"") != std::string::npos);
    ASSERT_TRUE(json.find("\"key2\":123") != std::string::npos);
    ASSERT_TRUE(json.find("\"key3\":true") != std::string::npos);
    ASSERT_TRUE(json.find("}") != std::string::npos);
    
    return true;
}

TEST(test_json_number_format) {
    int pid = 1234;
    std::string json = "{\"pid\":" + std::to_string(pid) + "}";
    
    ASSERT_TRUE(json.find("1234") != std::string::npos);
    return true;
}

void register_json_tests() {
    REGISTER_TEST(test_json_escape_quote);
    REGISTER_TEST(test_json_escape_backslash);
    REGISTER_TEST(test_json_escape_newline);
    REGISTER_TEST(test_json_escape_tab);
    REGISTER_TEST(test_json_escape_mixed);
    REGISTER_TEST(test_json_no_escape_needed);
    REGISTER_TEST(test_timestamp_format);
    REGISTER_TEST(test_build_json_object);
    REGISTER_TEST(test_json_number_format);
}