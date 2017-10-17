#ifndef NNTPCHAN_FRONTEND_HPP
#define NNTPCHAN_FRONTEND_HPP
#include <experimental/filesystem>
#include <memory>
#include <string>
namespace nntpchan
{

namespace fs = std::experimental::filesystem;
/** @brief nntpchan frontend ui interface */
class Frontend
{
public:
  /** @brief process an inbound message stored at fpath that we have accepted. */
  virtual void ProcessNewMessage(const fs::path &fpath) = 0;

  /** @brief return true if we take posts in a newsgroup */
  virtual bool AcceptsNewsgroup(const std::string &newsgroup) = 0;

  /** @brief return true if we will accept a message given its message-id */
  virtual bool AcceptsMessage(const std::string &msgid) = 0;
};

typedef std::unique_ptr<Frontend> Frontend_ptr;
}

#endif
