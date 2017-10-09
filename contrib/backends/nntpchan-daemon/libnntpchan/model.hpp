#ifndef NNTPCHAN_MODEL_HPP
#define NNTPCHAN_MODEL_HPP
#include <map>
#include <set>
#include <string>
#include <tuple>
#include <vector>

namespace nntpchan
{
  namespace model
  {
    // MIME Header
    typedef std::map<std::string, std::set<std::string> > PostHeader;
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


    static inline std::string & GetFilename(PostAttachment & att)
    {
      return std::get<0>(att);
    }
    
    static inline std::string & GetHexDigest(PostAttachment & att)
    {
      return std::get<1>(att);
    }
    
    static inline std::string & GetThumbnail(PostAttachment & att)
    {
      return std::get<2>(att);
    }
    
    static inline PostHeader & GetHeader(Post & post)
    {
      return std::get<0>(post);
    }
    
    static inline PostBody & GetBody(Post & post)
    {
      return std::get<1>(post);
    }
    
    static inline Attachments & GetAttachments(Post & post)
    {
      return std::get<2>(post);
    }
  }
}

#endif
