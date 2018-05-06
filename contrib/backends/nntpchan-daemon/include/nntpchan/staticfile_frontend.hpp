#ifndef NNTPCHAN_STATICFILE_FRONTEND_HPP
#define NNTPCHAN_STATICFILE_FRONTEND_HPP
#include "frontend.hpp"
#include "message.hpp"
#include "model.hpp"
#include "template_engine.hpp"
#include <experimental/filesystem>

namespace nntpchan
{

namespace fs = std::experimental::filesystem;

class StaticFileFrontend : public Frontend
{
public:
  StaticFileFrontend(TemplateEngine *tmpl, const std::string &templateDir, const std::string &outDir, uint32_t pages);

  virtual ~StaticFileFrontend();

  virtual void ProcessNewMessage(const fs::path &fpath);
  virtual bool AcceptsNewsgroup(const std::string &newsgroup);
  virtual bool AcceptsMessage(const std::string &msgid);

private:
  MessageDB_ptr m_MessageDB;
  TemplateEngine_ptr m_TemplateEngine;
  fs::path m_TemplateDir;
  fs::path m_OutDir;
  uint32_t m_Pages;
};
}

#endif
