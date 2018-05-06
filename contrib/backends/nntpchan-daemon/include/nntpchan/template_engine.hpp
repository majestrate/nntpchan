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
  virtual ~TemplateEngine() {};
  virtual bool WriteBoardPage(const nntpchan::model::BoardPage & page, const FileHandle_ptr &out) = 0;
  virtual bool WriteThreadPage(const nntpchan::model::Thread & thread, const FileHandle_ptr &out) = 0;
};



typedef std::unique_ptr<TemplateEngine> TemplateEngine_ptr;

TemplateEngine * CreateTemplateEngine(const std::string &dialect);
}

#endif
