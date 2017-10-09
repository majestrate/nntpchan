#ifndef NNTPCHAN_TEMPLATE_ENGINE_HPP
#define NNTPCHAN_TEMPLATE_ENGINE_HPP
#include "file_handle.hpp"
#include <any>
#include <map>
#include <memory>
#include <string>

namespace nntpchan
{

  struct TemplateEngine
  {
    using Args_t = std::map<std::string, std::any>;
    virtual bool WriteTemplate(const std::string & template_fname, const Args_t & args, const FileHandle_ptr & out) = 0;
  };

  TemplateEngine * CreateTemplateEngine(const std::string & dialect);

  typedef std::unique_ptr<TemplateEngine> TemplateEngine_ptr;
}

#endif
