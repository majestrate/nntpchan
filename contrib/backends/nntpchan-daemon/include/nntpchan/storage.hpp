#ifndef NNTPCHAN_STORAGE_HPP
#define NNTPCHAN_STORAGE_HPP

#include <experimental/filesystem>
#include <string>
#include "file_handle.hpp"

namespace nntpchan
{

  namespace fs = std::experimental::filesystem;
  
  class ArticleStorage
  {
  public:
    ArticleStorage();
    ArticleStorage(const fs::path & fpath);
    ~ArticleStorage();

    void SetPath(const fs::path & fpath);

    FileHandle_ptr OpenWrite(const std::string & msgid);
    FileHandle_ptr OpenRead(const std::string & msgid);

    /**
       return true if we should accept a new message give its message id
     */
    bool Accept(const std::string & msgid);
    
  private:

    fs::path MessagePath(const std::string & msgid);

    fs::path basedir;

  };
}


#endif
