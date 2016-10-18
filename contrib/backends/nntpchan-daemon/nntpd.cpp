#include "ini.hpp"

#include "storage.hpp"
#include "nntp_server.hpp"
#include "event.hpp"
#include "exec_frontend.hpp"

#include <vector>
#include <string>


int main(int argc, char * argv[]) {
  if (argc != 2) {
    std::cerr << "usage: " << argv[0] << " config.ini" << std::endl;
    return 1;
  }

  nntpchan::Mainloop loop;
  
  nntpchan::NNTPServer nntp(loop);
  
  std::string fname(argv[1]);

  std::ifstream i(fname);

  if(i.is_open()) {
    INI::Parser conf(i);
    
    std::vector<std::string> requiredSections = {"nntp", "storage"};
    
    auto & level = conf.top();
    
    for ( const auto & section : requiredSections ) {
      if(level.sections.find(section) == level.sections.end()) {
        std::cerr << "config file " << fname << " does not have required section: ";
        std::cerr << section << std::endl;
        return 1;
      }
    }

    auto & storeconf = level.sections["storage"].values;

    if (storeconf.find("path") == storeconf.end()) {
      std::cerr << "storage section does not have 'path' value" << std::endl;
      return 1;
    }

    nntp.SetStoragePath(storeconf["path"]);
    
    auto & nntpconf = level.sections["nntp"].values;

    if (nntpconf.find("bind") == nntpconf.end()) {
      std::cerr << "nntp section does not have 'bind' value" << std::endl;
      return 1;
    }

    if (nntpconf.find("authdb") != nntpconf.end()) {
      nntp.SetLoginDB(nntpconf["authdb"]);
    }

    if ( level.sections.find("frontend") != level.sections.end()) {
      // frontend enabled
      auto & frontconf = level.sections["frontend"].values;
      if (frontconf.find("type") == frontconf.end()) {
        std::cerr << "frontend section provided but 'type' value not provided" << std::endl;
        return 1;
      }
      auto ftype = frontconf["type"];
      if (ftype == "exec") {
        if (frontconf.find("exec") == frontconf.end()) {
          std::cerr << "exec frontend specified but no 'exec' value provided" << std::endl;
          return 1;
        }
        nntp.SetFrontend(new nntpchan::ExecFrontend(frontconf["exec"]));
      } else {
        std::cerr << "unknown frontend type '" << ftype << "'" << std::endl;
      }
          
    }
    
    auto & a = nntpconf["bind"];

    try {
      nntp.Bind(a);  
    } catch ( std::exception & ex ) {
      std::cerr << "failed to bind: " << ex.what() << std::endl;
      return 1;
    }
    try {
      std::cerr << "run mainloop" << std::endl;
      loop.Run();
    } catch ( std::exception & ex ) {
      std::cerr << "exception in mainloop: " << ex.what() << std::endl;
      return 2;
    }
    
  } else {
    std::cerr << "failed to open " << fname << std::endl;
    return 1;
  }
    
  
}
