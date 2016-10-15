#include "base64.hpp"
#include "crypto.hpp"

#include <string>
#include <iostream>

static void print_help(const std::string & exename)
{
  std::cout << "usage: " << exename << " [help|passwd|genconf]" << std::endl;
}

int main(int argc, char * argv[])
{
  if(argc == 1) {
    print_help(argv[0]);
    return 1;
  }
}
