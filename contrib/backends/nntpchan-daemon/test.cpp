#include "exec_frontend.hpp"
#include "sanitize.hpp"
#include <cassert>
#include <iostream>



int main(int , char * [])
{
  nntpchan::Frontend_ptr f (new nntpchan::ExecFrontend("./contrib/nntpchan.sh"));
  assert(f->AcceptsMessage("<test@server>"));
  assert(f->AcceptsNewsgroup("overchan.test"));
  assert(nntpchan::IsValidMessageID("<test@test>"));
  assert(!nntpchan::IsValidMessageID("asd"));
  std::cout << "all good" << std::endl;
}
