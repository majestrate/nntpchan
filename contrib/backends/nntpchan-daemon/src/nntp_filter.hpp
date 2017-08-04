#ifndef NNTP_FILTER_HPP
#define NNTP_FILTER_HPP

#include <iostream>
#include <crypto.hpp>

namespace nntpchan
{

  template<class CharT>
  struct MessageFilter : public std::basic_ostream<CharT>
  {
    using Base = std::basic_streambuf<CharT>;
    using char_type = Base::char_type;
    using int_type = Base::int_type;

    virtual int_type overflow(int_type ch);
  protected:
    virtual void xsputn(const char_type * data, std::streamsize sz);



  };
}

#endif
