#include "exec_frontend.hpp"
#include <cassert>
#include <iostream>



int main(int , char * [])
{
  nntpchan::Frontend * f = new nntpchan::ExecFrontend("./contrib/nntpchan.sh");
  assert(f->AcceptsMessage("<test@server>"));
  assert(f->AcceptsNewsgroup("overchan.test"));
  std::cout << "all good" << std::endl;
}
