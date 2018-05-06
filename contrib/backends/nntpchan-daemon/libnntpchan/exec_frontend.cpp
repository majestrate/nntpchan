#include <cstring>
#include <errno.h>
#include <iostream>
#include <nntpchan/exec_frontend.hpp>
#include <sys/wait.h>
#include <unistd.h>

namespace nntpchan
{
ExecFrontend::ExecFrontend(const std::string &fname, char * const* env) : m_Environ(env), m_exec(fname) {}

ExecFrontend::~ExecFrontend() {}

void ExecFrontend::ProcessNewMessage(const fs::path &fpath) { Exec({"post", fpath}); }

bool ExecFrontend::AcceptsNewsgroup(const std::string &newsgroup) { return Exec({"newsgroup", newsgroup}) == 0; }

bool ExecFrontend::AcceptsMessage(const std::string &msgid) { return Exec({"msgid", msgid}) == 0; }

int ExecFrontend::Exec(std::deque<std::string> args)
{
  // set up arguments
  const char **cargs = new char const *[args.size() + 2];
  std::size_t l = 0;
  cargs[l++] = m_exec.c_str();
  while (args.size())
  {
    cargs[l++] = args.front().c_str();
    args.pop_front();
  }
  cargs[l] = 0;
  int retcode = 0;
  pid_t child = fork();
  if (child)
  {
    waitpid(child, &retcode, 0);
  }
  else
  {
    int r = execvpe(m_exec.c_str(), (char *const *)cargs, m_Environ);
    if (r == -1)
    {
      std::cout << strerror(errno) << std::endl;
      exit(errno);
    }
    else
      exit(r);
  }
  return retcode;
}
}
