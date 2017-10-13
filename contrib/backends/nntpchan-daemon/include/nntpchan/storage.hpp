#ifndef NNTPCHAN_STORAGE_HPP
#define NNTPCHAN_STORAGE_HPP

#include <experimental/filesystem>
#include <string>
#include "file_handle.hpp"
#include "message.hpp"

namespace nntpchan
{

  namespace fs = std::experimental::filesystem;
  
  class ArticleStorage : public MessageDB
  {
  public:
    ArticleStorage();
    ArticleStorage(const fs::path & fpath);
    ~ArticleStorage();

    FileHandle_ptr OpenWrite(const std::string & msgid) const;
    FileHandle_ptr OpenRead(const std::string & msgid) const;

    /**
       return true if we should accept a new message give its message id
     */
    bool Accept(const std::string & msgid) const;

    bool LoadBoardPage(BoardPage & board, const std::string & newsgroup, uint32_t perpage, uint32_t page) const;
    bool FindThreadByHash(const std::string & hashhex, std::string & msgid) const;
    bool LoadThread(Thread & thread, const std::string & rootmsgid) const;

    /** ensure symlinks are formed for this article by message id */
    void EnsureSymlinks(const std::string & msgid) const;
    
  private:
    void SetPath(const fs::path & fpath);

    fs::path MessagePath(const std::string & msgid) const;

    bool init_skiplist(const std::string & subdir) const;

    fs::path skiplist_root(const std::string & name) const;
    
    fs::path basedir;

  };

  typedef std::unique_ptr<ArticleStorage> ArticleStorage_ptr;
}


#endif
