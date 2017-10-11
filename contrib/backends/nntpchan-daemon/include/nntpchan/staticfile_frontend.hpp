#ifndef NNTPCHAN_STATICFILE_FRONTEND_HPP
#define NNTPCHAN_STATICFILE_FRONTEND_HPP
#include "frontend.hpp"
#include "template_engine.hpp"
#include "model.hpp"
#include <experimental/filesystem>

namespace nntpchan
{

  namespace fs = std::experimental::filesystem;
  
  class StaticFileFrontend : public Frontend
  {
  public:

    StaticFileFrontend(TemplateEngine * tmpl, const std::string & templateDir, const std::string & outDir, uint32_t pages);
    
    ~StaticFileFrontend();

    void ProcessNewMessage(const fs::path & fpath);
    bool AcceptsNewsgroup(const std::string & newsgroup);
    bool AcceptsMessage(const std::string & msgid);

  private:

    typedef nntpchan::model::Thread Thread_t;

    typedef std::vector<Thread_t> Threads_t;

    typedef std::vector<Threads_t> BoardPage_t;
    
    BoardPage_t GetThreadsPaginated(const std::string & group, uint32_t perpage, uint32_t pages);
    
  private:
    
    TemplateEngine_ptr m_TemplateEngine;
    fs::path m_TemplateDir;
    fs::path m_OutDir;
    uint32_t m_Pages;
  };
}

#endif
