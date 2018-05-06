#include <any>
#include <iostream>
#include <nntpchan/file_handle.hpp>
#include <nntpchan/mime.hpp>
#include <nntpchan/sanitize.hpp>
#include <nntpchan/sha1.hpp>
#include <nntpchan/staticfile_frontend.hpp>
#include <set>
#include <sstream>

namespace nntpchan
{

StaticFileFrontend::StaticFileFrontend(TemplateEngine *tmpl, const std::string &templateDir, const std::string &outDir,
                                       uint32_t pages)
    : m_TemplateEngine(tmpl), m_TemplateDir(templateDir), m_OutDir(outDir), m_Pages(pages)
{
}

void StaticFileFrontend::ProcessNewMessage(const fs::path &fpath)
{
  std::clog << "process message " << fpath << std::endl;
  auto file = OpenFile(fpath, eRead);
  if (file)
  {
    // read header
    RawHeader header;
    if (!ReadHeader(file, header))
    {
      std::clog << "failed to read mime header" << std::endl;
      return;
    }

    // read body

    auto findMsgidFunc = [](const std::pair<std::string, std::string> &item) -> bool {
      auto lower = ToLower(item.first);
      return (lower == "message-id") || (lower == "messageid");
    };

    auto msgid_itr = std::find_if(header.begin(), header.end(), findMsgidFunc);
    if (msgid_itr == std::end(header))
    {
      std::clog << "no message id for file " << fpath << std::endl;
      return;
    }

    std::string msgid = StripWhitespaces(msgid_itr->second);

    if (!IsValidMessageID(msgid))
    {
      std::clog << "invalid message-id: " << msgid << std::endl;
      return;
    }

    std::string rootmsgid;

    auto findReferences = [](const std::pair<std::string, std::string> &item) -> bool {
      auto lower = ToLower(item.first);
      return lower == "references";
    };

    auto references_itr = std::find_if(header.begin(), header.end(), findReferences);
    if (references_itr == std::end(header) || StripWhitespaces(references_itr->second).size() == 0)
    {
      rootmsgid = msgid;
    }
    else
    {
      const auto &s = references_itr->second;
      auto checkfunc = [](unsigned char ch) -> bool { return std::isspace(ch) || std::iscntrl(ch); };
      if (std::count_if(s.begin(), s.end(), checkfunc))
      {
        /** split off first element */
        auto idx = std::find_if(s.begin(), s.end(), checkfunc);
        rootmsgid = s.substr(0, s.find(*idx));
      }
      else
      {
        rootmsgid = references_itr->second;
      }
    }

    std::string rootmsgid_hash = sha1_hex(rootmsgid);

    std::set<std::string> newsgroups_list;

    auto findNewsgroupsFunc = [](const std::pair<std::string, std::string> &item) -> bool {
      return ToLower(item.first) == "newsgroups";
    };

    auto group = std::find_if(header.begin(), header.end(), findNewsgroupsFunc);
    if (group == std::end(header))
    {
      std::clog << "no newsgroups header" << std::endl;
      return;
    }
    std::istringstream input(group->second);

    std::string newsgroup;
    while (std::getline(input, newsgroup, ' '))
    {
      if (IsValidNewsgroup(newsgroup))
        newsgroups_list.insert(newsgroup);
    }

    fs::path threadFilePath = m_OutDir / fs::path("thread-" + rootmsgid_hash + ".html");
    nntpchan::model::Thread thread;

    if (!m_MessageDB)
    {
      std::clog << "no message database" << std::endl;
      return;
    }

    if (!m_MessageDB->LoadThread(thread, rootmsgid))
    {
      std::clog << "cannot find thread with root " << rootmsgid << std::endl;
      return;
    }
    if (m_TemplateEngine)
    {
      FileHandle_ptr out = OpenFile(threadFilePath, eWrite);
      if (!out || !m_TemplateEngine->WriteThreadPage(thread, out))
      {
        std::clog << "failed to write " << threadFilePath << std::endl;
        return;
      }
    }
    nntpchan::model::BoardPage page;
    for (const auto &name : newsgroups_list)
    {
      uint32_t pageno = 0;
      while (pageno < m_Pages)
      {
        page.threads.clear();
        if (!m_MessageDB->LoadBoardPage(page, name, 10, m_Pages))
        {
          std::clog << "cannot load board page " << pageno << " for " << name << std::endl;
          break;
        }
        fs::path boardPageFilename(name + "-" + std::to_string(pageno) + ".html");
        if (m_TemplateEngine)
        {
          fs::path outfile = m_OutDir / boardPageFilename;
          FileHandle_ptr out = OpenFile(outfile, eWrite);
          if (out)
            m_TemplateEngine->WriteBoardPage(page, out);
          else
            std::clog << "failed to open board page " << outfile << std::endl;
        }

        ++pageno;
      }
    }
  }
}

bool StaticFileFrontend::AcceptsNewsgroup(const std::string &newsgroup) { return IsValidNewsgroup(newsgroup); }

bool StaticFileFrontend::AcceptsMessage(const std::string &msgid) { return IsValidMessageID(msgid); }
}
