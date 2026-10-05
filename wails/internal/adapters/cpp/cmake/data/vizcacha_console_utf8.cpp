// Added by VizcachaIDE to every C++ program it builds on Windows, so accents work in the console:
// - output: the console shows UTF-8 ("¿Cómo te llamas?" instead of "┐C├│mo");
// - input: std::cin reads what is typed with ReadConsoleW and gets UTF-8 ("Ñandú"), because a
//   console in UTF-8 drops every non-ASCII character typed when it is read byte by byte.
// Nothing changes when the input is not a console (a file or a pipe).
#include <windows.h>

#include <iostream>
#include <streambuf>
#include <string>

namespace {

class ConsoleInput : public std::streambuf {
public:
    explicit ConsoleInput(HANDLE console) : console_(console) {}

protected:
    int_type underflow() override {
        if (gptr() < egptr()) {
            return traits_type::to_int_type(*gptr());
        }
        wchar_t wide[1024];
        DWORD read = 0;
        if (!ReadConsoleW(console_, wide, 1024, &read, nullptr) || read == 0) {
            return traits_type::eof();
        }
        int size = WideCharToMultiByte(CP_UTF8, 0, wide, static_cast<int>(read), nullptr, 0, nullptr, nullptr);
        text_.assign(static_cast<size_t>(size), '\0');
        WideCharToMultiByte(CP_UTF8, 0, wide, static_cast<int>(read), &text_[0], size, nullptr, nullptr);
        std::string::size_type at;
        while ((at = text_.find("\r\n")) != std::string::npos) {
            text_.erase(at, 1);  // the console ends lines with \r\n; std::getline wants \n
        }
        setg(&text_[0], &text_[0], &text_[0] + text_.size());
        return traits_type::to_int_type(*gptr());
    }

private:
    HANDLE console_;
    std::string text_;
};

struct VizcachaUtf8Console {
    VizcachaUtf8Console() {
        SetConsoleOutputCP(CP_UTF8);
        HANDLE input = GetStdHandle(STD_INPUT_HANDLE);
        DWORD mode = 0;
        if (input != INVALID_HANDLE_VALUE && GetConsoleMode(input, &mode)) {
            static ConsoleInput console(input);
            std::cin.rdbuf(&console);
        }
    }
} vizcachaUtf8Console;

}  // namespace
