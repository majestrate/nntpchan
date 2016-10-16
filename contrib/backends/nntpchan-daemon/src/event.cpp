#include "event.hpp"
#include <cassert>

namespace nntpchan
{
  Mainloop::Mainloop()
  {
    m_loop = uv_default_loop();
    assert(uv_loop_init(m_loop) == 0);
  }

  Mainloop::~Mainloop()
  {
    uv_loop_close(m_loop);
  }

  void Mainloop::Stop()
  {
    uv_stop(m_loop);
  }

  void Mainloop::Run(uv_run_mode mode)
  {
    uv_run(m_loop, mode);
  }
  
}
