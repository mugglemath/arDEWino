#include"Utils.h"
#include <regex>
#include <string>

bool isValidFloatFormat(const std::string& input) {
    std::regex floatPattern(R"(\b(\d{2}\.\d{2},\d{2}\.\d{2})\b)");
    return std::regex_match(input, floatPattern);
}

bool isValidSingleCharacter(const std::string& input) {
    return input.length() == 1 && std::isalnum(input[0]);
}
