#include <iostream>
#include <mstch/mstch.hpp>
#include <nntpchan/sanitize.hpp>
#include <nntpchan/template_engine.hpp>
#include <sstream>

namespace nntpchan
{

template <class... Ts> struct overloaded : Ts...
{
  using Ts::operator()...;
};
template <class... Ts> overloaded(Ts...)->overloaded<Ts...>;

namespace mustache = mstch;

static mustache::map post_to_map(const nntpchan::model::Post &post)
{
  mustache::map m;
  mustache::array attachments;
  mustache::map h;

  for (const auto &att : nntpchan::model::GetAttachments(post))
  {
    mustache::map a;
    a["filename"] = nntpchan::model::GetFilename(att);
    a["hexdigest"] = nntpchan::model::GetHexDigest(att);
    a["thumbnail"] = nntpchan::model::GetThumbnail(att);
    attachments.push_back(a);
  }

  for (const auto &item : nntpchan::model::GetHeader(post))
  {
    mustache::array vals;
    for (const auto &v : item.second)
      vals.push_back(v);
    h[item.first] = vals;
  }

  m["attachments"] = attachments;
  m["message"] = nntpchan::model::GetBody(post);
  m["header"] = h;
  return m;
}

static mustache::map thread_to_map(const nntpchan::model::Thread &t)
{
  mustache::map thread;
  mustache::array posts;
  for (const auto &post : t)
  {
    posts.push_back(post_to_map(post));
  }
  auto &opHeader = nntpchan::model::GetHeader(t[0]);
  thread["title"] = nntpchan::model::HeaderIFind(opHeader, "subject", "None")[0];
  thread["posts"] = posts;
  return thread;
}

struct MustacheTemplateEngine : public TemplateEngine
{
  struct Impl
  {

    Impl(const std::map<std::string, std::string> &partials) : m_partials(partials) {}

    bool ParseTemplate(const FileHandle_ptr &in)
    {
      std::stringstream str;
      std::string line;
      while (std::getline(*in, line))
        str << line << "\n";
      m_tmplString = str.str();
      return in->eof();
    }

    bool RenderFile(const Args_t &args, const FileHandle_ptr &out)
    {
      mustache::map obj;
      for (const auto &item : args)
      {
        std::visit(overloaded{[&obj, item](const nntpchan::model::Model &m) {
                                std::visit(overloaded{[&obj, item](const nntpchan::model::BoardPage &p) {
                                                        mustache::array threads;
                                                        for (const auto &thread : p)
                                                        {
                                                          threads.push_back(thread_to_map(thread));
                                                        }
                                                        obj[item.first] = threads;
                                                      },
                                                      [&obj, item](const nntpchan::model::Thread &t) {
                                                        obj[item.first] = thread_to_map(t);
                                                      }},
                                           m);
                              },
                              [&obj, item](const std::string &str) { obj[item.first] = str; }},
                   item.second);
      }

      std::string str = mustache::render(m_tmplString, obj);
      out->write(str.c_str(), str.size());
      out->flush();
      return !out->fail();
    }

    std::string m_tmplString;
    const std::map<std::string, std::string> &m_partials;
  };

  virtual bool WriteTemplate(const fs::path &fpath, const Args_t &args, const FileHandle_ptr &out)
  {
    auto templFile = OpenFile(fpath, eRead);
    if (!templFile)
    {
      std::clog << "no such template at " << fpath << std::endl;
      return false;
    }

    std::map<std::string, std::string> partials;
    if (!LoadPartials(fpath.parent_path(), partials))
    {
      std::clog << "failed to load partials" << std::endl;
      return false;
    }

    Impl impl(partials);
    if (impl.ParseTemplate(templFile))
    {
      return impl.RenderFile(args, out);
    }

    std::clog << "failed to parse template " << fpath << std::endl;
    return false;
  }

  bool LoadPartials(fs::path dir, std::map<std::string, std::string> &partials)
  {
    const auto partial_files = {"header", "footer"};
    for (const auto &fname : partial_files)
    {
      auto file = OpenFile(dir / fs::path(fname + std::string(".html")), eRead);
      if (!file)
      {
        std::clog << "no such partial: " << fname << std::endl;
        return false;
      }
      std::string line;
      std::stringstream input;
      while (std::getline(*file, line))
        input << line << "\n";
      partials[fname] = input.str();
    }
    return true;
  }
};

TemplateEngine *CreateTemplateEngine(const std::string &dialect)
{
  auto d = ToLower(dialect);
  if (d == "mustache")
    return new MustacheTemplateEngine;
  else
    return nullptr;
}
}
