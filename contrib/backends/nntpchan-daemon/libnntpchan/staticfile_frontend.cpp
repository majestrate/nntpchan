#include "staticfile_frontend.hpp"
#include "file_handle.hpp"
#include "sanitize.hpp"
#include "mime.hpp"
#include "sha1.hpp"
#include <any>
#include <iostream>
#include <set>
#include <sstream>

namespace nntpchan
{


  StaticFileFrontend::StaticFileFrontend(TemplateEngine * tmpl, const std::string & templateDir, const std::string & outDir, uint32_t pages) :
    m_TemplateEngine(tmpl),
    m_TemplateDir(templateDir),
    m_OutDir(outDir),
    m_Pages(pages)
  {
  }

  void StaticFileFrontend::ProcessNewMessage(const fs::path & fpath)
  {
    std::clog << "process message " << fpath << std::endl;
    auto file = OpenFile(fpath, eRead);
    if(file)
    {
      // read header
      RawHeader header;
      if(!ReadHeader(file, header))
      {
        std::clog << "failed to read mime header" << std::endl;
        return;
      }

      // read body

      // render templates
      if(m_TemplateEngine)
      {
        std::map<std::string, std::any> thread_args;

        auto findMsgidFunc = [](const std::pair<std::string, std::string> & item) -> bool {
          auto lower = ToLower(item.first);
          return (lower == "message-id") || (lower == "messageid");
        };
        
        auto msgid = std::find_if(header.begin(), header.end(), findMsgidFunc);
        
        std::string msgid_hash = sha1_hex(msgid->second);
        
        fs::path threadFilePath = m_OutDir / fs::path("thread-" + msgid_hash + ".html");
        FileHandle_ptr out = OpenFile(threadFilePath, eWrite);
        if(!m_TemplateEngine->WriteTemplate("thread.mustache", thread_args, out))
        {
          std::clog << "failed to write " << threadFilePath << std::endl;
          return;
        }

        std::set<std::string> newsgroups_list;

        auto findNewsgroupsFunc = [](const std::pair<std::string, std::string> & item) -> bool
        {
          return ToLower(item.first) == "newsgroups";
        };

        auto group = std::find_if(header.begin(), header.end(), findNewsgroupsFunc);
        if(group == std::end(header))
        {
          std::clog << "no newsgroups header" << std::endl;
          return;
        }
        std::istringstream input(group->second);
       
        std::string newsgroup;
        while(std::getline(input, newsgroup, ' '))
        {
          newsgroups_list.insert(NNTPSanitize(newsgroup));
        }

        for(const auto & name : newsgroups_list)
        {
          auto board = GetThreadsPaginated(name, 10, m_Pages);
          uint32_t pageno = 0;
          for(Threads_t threads : board)
          {
            std::map<std::string, std::any> board_args;
            board_args["group"] = std::make_any<std::string>(name);
            board_args["pageno"] = std::make_any<uint32_t>(pageno);
            board_args["threads"] = std::make_any<Threads_t>(threads);
            
            fs::path boardPageFilename(newsgroup + "-" + std::to_string(pageno) + ".html");
            out = OpenFile(m_OutDir / boardPageFilename, eWrite);
            m_TemplateEngine->WriteTemplate("board.mustache", board_args, out);

            ++pageno;
          }
        }
      }
    }
  }
  
  StaticFileFrontend::BoardPage_t StaticFileFrontend::GetThreadsPaginated(const std::string & group, uint32_t perpage, uint32_t pages)
  {
    return {};
  }
}
