#include <iostream>
#include <nntpchan/sanitize.hpp>
#include <nntpchan/template_engine.hpp>
#include <sstream>

namespace nntpchan
{
  struct StdTemplateEngine : public TemplateEngine
  {
    struct RenderContext 
    {

      bool Load(const fs::path & path)
      {
        // clear out previous data
        m_Data.clear();
        // open file
        std::ifstream f;
        f.open(path);
        if(f.is_open())
        {
          for(std::string line; std::getline(f, line, '\n');)
          {
            m_Data += line + "\n";
          }
          return true;
        }
        else 
          return false;
      }

      virtual bool Render(const FileHandle_ptr & out) const = 0;

      std::string m_Data;
    };

    struct BoardRenderContext : public RenderContext
    {
      const nntpchan::model::BoardPage & m_Page;

      BoardRenderContext(const nntpchan::model::BoardPage & page) : m_Page(page) {};

      virtual bool Render(const FileHandle_ptr & out) const
      {
        *out << m_Data;
        return false;
      }
    };

    struct ThreadRenderContext : public RenderContext
    {
      const nntpchan::model::Thread & m_Thread;

      ThreadRenderContext(const nntpchan::model::Thread & thread) : m_Thread(thread) {};

      virtual bool Render(const FileHandle_ptr & out) const 
      {
        *out << m_Data;
        return false;
      }

    };

    bool WriteBoardPage(const nntpchan::model::BoardPage & page, const FileHandle_ptr & out)
    {
      BoardRenderContext ctx(page);
      if(ctx.Load("board.html"))
      {
        return ctx.Render(out);
      }
      return false;
    }

    bool WriteThreadPage(const nntpchan::model::Thread & thread, const FileHandle_ptr & out)
    {
      ThreadRenderContext ctx(thread);
      if(ctx.Load("thread.html"))
      { 
        return ctx.Render(out);
      }
      return false;
    }
  };

TemplateEngine *CreateTemplateEngine(const std::string &dialect)
{
  auto d = ToLower(dialect);
  if (d == "std")
    return new StdTemplateEngine;
  else
    return nullptr;
}
}
