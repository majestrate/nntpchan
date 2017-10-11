#include <nntpchan/template_engine.hpp>
#include <nntpchan/sanitize.hpp>
#include <iostream>

namespace nntpchan
{
  
  struct MustacheTemplateEngine : public TemplateEngine
  {
    struct Impl
    {

      bool ParseTemplate(const FileHandle_ptr & in)
      {
        return true;
      }
      
      bool RenderFile(const Args_t & args, const FileHandle_ptr & out)
      {
        return true;
      }

    };

    virtual bool WriteTemplate(const fs::path & fpath, const Args_t & args, const FileHandle_ptr & out)
    {
      auto templFile = OpenFile(fpath, eRead);
      if(!templFile)
      {
        std::clog << "no such template at " << fpath << std::endl;
        return false;
      }
      auto impl = std::make_unique<Impl>();
      if(impl->ParseTemplate(templFile))
        return impl->RenderFile(args, out);

      std::clog << "failed to parse template " << fpath << std::endl;
      return false;
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
