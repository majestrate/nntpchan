#ifndef NNTPCHAN_EVENT_HPP
#define NNTPCHAN_EVENT_HPP
#include <uv.h>

namespace nntpchan
{
  class Mainloop
  {
  public:

    Mainloop();
    ~Mainloop();

    operator uv_loop_t * () const { return m_loop; }

    void Run(uv_run_mode mode = UV_RUN_DEFAULT);
    void Stop();
    
  private:

    uv_loop_t * m_loop;
    
  };
}

#endif
