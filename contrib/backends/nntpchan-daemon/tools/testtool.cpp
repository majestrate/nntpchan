#include <cassert>
#include <iostream>
#include <nntpchan/exec_frontend.hpp>
#include <nntpchan/sanitize.hpp>

int main(int, char *[])
{
  nntpchan::Frontend_ptr f(new nntpchan::ExecFrontend("./contrib/nntpchan.sh"));
  assert(nntpchan::IsValidMessageID("<a28a71493831188@web.oniichan.onion>"));
  assert(f->AcceptsNewsgroup("overchan.test"));
  std::cout << "all good" << std::endl;
}
