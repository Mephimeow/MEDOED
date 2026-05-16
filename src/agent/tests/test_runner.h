#ifndef TEST_RUNNER_H
#define TEST_RUNNER_H

#include <iostream>
#include <string>
#include <vector>
#include <functional>
#include <chrono>

#define ASSERT_EQ(expected, actual) \
    if ((expected) != (actual)) { \
        std::cerr << "FAIL: " << __FILE__ << ":" << __LINE__ << " - Expected " << (expected) << ", got " << (actual) << std::endl; \
        return false; \
    }

#define ASSERT_TRUE(condition) \
    if (!(condition)) { \
        std::cerr << "FAIL: " << __FILE__ << ":" << __LINE__ << " - Expected true, got false" << std::endl; \
        return false; \
    }

#define ASSERT_FALSE(condition) \
    if ((condition)) { \
        std::cerr << "FAIL: " << __FILE__ << ":" << __LINE__ << " - Expected false, got true" << std::endl; \
        return false; \
    }

#define ASSERT_NE(expected, actual) \
    if ((expected) == (actual)) { \
        std::cerr << "FAIL: " << __FILE__ << ":" << __LINE__ << " - Expected not " << (expected) << std::endl; \
        return false; \
    }

class TestRunner {
public:
    static TestRunner& instance() {
        static TestRunner runner;
        return runner;
    }

    void add_test(const std::string& name, std::function<bool()> test) {
        tests_.push_back({name, test});
    }

    int run_all() {
        int passed = 0;
        int failed = 0;
        std::cout << "Running " << tests_.size() << " tests...\n\n";

        for (const auto& test : tests_) {
            auto start = std::chrono::high_resolution_clock::now();
            bool result = test.second();
            auto end = std::chrono::high_resolution_clock::now();
            auto duration = std::chrono::duration_cast<std::chrono::microseconds>(end - start);

            if (result) {
                std::cout << "[PASS] " << test.first << " (" << duration.count() << "μs)\n";
                passed++;
            } else {
                std::cout << "[FAIL] " << test.first << "\n";
                failed++;
            }
        }

        std::cout << "\n" << passed << " passed, " << failed << " failed\n";
        return failed;
    }

    void reset() {
        tests_.clear();
    }

private:
    TestRunner() = default;
    std::vector<std::pair<std::string, std::function<bool()>>> tests_;
};

#define TEST(name) bool test_##name()
#define REGISTER_TEST(name) TestRunner::instance().add_test(#name, test_##name)

#define RUN_TESTS() TestRunner::instance().run_all()
#define RESET_TESTS() TestRunner::instance().reset()

#endif