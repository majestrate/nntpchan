#include "io.hpp"

namespace nntpchan
{
  namespace io
  {

    template<class CharT = char>
    struct MultiWriter : public std::basic_streambuf<CharT> , public std::basic_ostream<CharT>
    {
      using Base = std::basic_streambuf<CharT>;
      using char_type = typename Base::char_type;
      using int_type = typename Base::int_type;

      std::vector<std::basic_ostream<CharT> *> writers;

      int_type overflow(int_type ch)
      {
        return Base::overflow(ch);
      }

      virtual std::streamsize xputn(const char_type * data, std::streamsize sz)
      {
        for(const auto & itr : writers)
          itr->write(data, sz);
        return sz;
      }

      virtual int sync()
      {
        for(const auto & itr : writers)
          itr->flush();
        return 0;
      }

      void AddWriter(std::basic_ostream<CharT> * wr)
      {
        writers.push_back(wr);
      }

      virtual ~MultiWriter()
      {
        for(const auto & itr : writers)
          delete itr;
      }

    };


    std::basic_ostream<char> * multiWriter(const std::vector<std::basic_ostream<char> *> & writers)
    {
      MultiWriter<char> * wr = new MultiWriter<char>;
      for(auto & itr : writers)
        wr->AddWriter(itr);
      return wr;
    }
  }
}
