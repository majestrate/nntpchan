#ifndef NNTPCHAN_TEMPLATE_ENGINE_HPP
#define NNTPCHAN_TEMPLATE_ENGINE_HPP
#include "file_handle.hpp"
#include "model.hpp"
#include <any>
#include <map>
#include <memory>
#include <string>

namespace nntpchan
{

  struct TemplateEngine
  {
    typedef std::map<std::string, std::variant<nntpchan::model::Model, std::string>> Args_t;
    virtual bool WriteTemplate(const fs::path & template_fpath, const Args_t & args, const FileHandle_ptr & out) = 0;
  };

  TemplateEngine * CreateTemplateEngine(const std::string & dialect);

  typedef std::unique_ptr<TemplateEngine> TemplateEngine_ptr;
}

#endif
