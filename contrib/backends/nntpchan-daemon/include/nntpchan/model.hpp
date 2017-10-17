#ifndef NNTPCHAN_MODEL_HPP
#define NNTPCHAN_MODEL_HPP
#include <algorithm>
#include <map>
#include <nntpchan/sanitize.hpp>
#include <set>
#include <string>
#include <tuple>
#include <variant>
#include <vector>

namespace nntpchan
{
namespace model
{
// MIME Header
typedef std::map<std::string, std::vector<std::string>> PostHeader;
// text post contents
typedef std::string PostBody;
// single file attachment, (orig_filename, hexdigest, thumb_filename)
typedef std::tuple<std::string, std::string, std::string> PostAttachment;
// all attachments on a post
typedef std::vector<PostAttachment> Attachments;
// a post (header, Post Text, Attachments)
typedef std::tuple<PostHeader, PostBody, Attachments> Post;
// a thread (many posts in post order)
typedef std::vector<Post> Thread;
// a board page is many threads in bump order
typedef std::vector<Thread> BoardPage;

static inline const std::string &GetFilename(const PostAttachment &att) { return std::get<0>(att); }

static inline const std::string &GetHexDigest(const PostAttachment &att) { return std::get<1>(att); }

static inline const std::string &GetThumbnail(const PostAttachment &att) { return std::get<2>(att); }

static inline const PostHeader &GetHeader(const Post &post) { return std::get<0>(post); }

static inline const PostBody &GetBody(const Post &post) { return std::get<1>(post); }

static inline const Attachments &GetAttachments(const Post &post) { return std::get<2>(post); }

static inline const std::string &HeaderIFind(const PostHeader &header, const std::string &val,
                                             const std::string &fallback)
{
  std::string ival = ToLower(val);
  auto itr = std::find_if(header.begin(), header.end(),
                          [ival](const auto &item) -> bool { return ToLower(item.first) == ival; });
  if (itr == std::end(header))
    return fallback;
  else
    return itr->second[0];
}

using Model = std::variant<Thread, BoardPage>;
}
}

#endif
