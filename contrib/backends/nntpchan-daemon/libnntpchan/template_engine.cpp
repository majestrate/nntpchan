#include "template_engine.hpp"
#include "sanitize.hpp"

namespace nntpchan
{
  
  struct MustacheTemplateEngine : public TemplateEngine
  {
    struct Impl
    {
      bool RenderFile(const std::string & fname, const Args_t & args, const FileHandle_ptr & out)
      {
        auto file = OpenFile(fname, eRead);
        
        
        return true;
      }
    };

    virtual bool WriteTemplate(const std::string & fname, const Args_t & args, const FileHandle_ptr & out)
    {
      auto impl = std::make_unique<Impl>();
      return impl->RenderFile(fname, args, out);
    }
  };

  TemplateEngine * CreateTemplateEngine(const std::string & dialect)
  {
    auto d = ToLower(dialect);
    if(d == "mustache")
      return new MustacheTemplateEngine;
    else
      return nullptr;
  }
}
