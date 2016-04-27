function enable_theme(prefix, name) {
  if (prefix && name) {
    var theme = document.getElementById("current_theme");
    if (theme) {
      theme.href = prefix + "static/"+ name + ".css";
      var st = get_storage();
      st.nntpchan_prefix = prefix;
      st.nntpchan_theme = name;
    }
  }
}

// apply themes
var st = get_storage();
enable_theme(st.nntpchan_prefix, st.nntpchan_theme);
