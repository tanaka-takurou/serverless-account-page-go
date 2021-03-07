var loginFormSubmit = function() {
	document.getElementById("redirect_uri").value = location.origin + "/profile";
	document.login.submit();
};
var logoutFormSubmit = function() {
	document.getElementById("logout_uri").value = location.origin;
	document.logout.submit();
};
var postRequest = function(url, data, headers, callback, onerror) {
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    headers:       headers,
    url:           url
  })
  .done(function(res) {
    callback(res);
  })
  .fail(function(e) {
    onerror(e);
  });
};
var userinfoCall = function(url) {
  var hashs = getHashs();
  if( !hashs.access_token || hashs.access_token.length < 1 ) {
    $("#warning").text("Error: Access token is not in correct format.").removeClass("hidden").addClass("visible");
	return;
  }
  var headers = {'Authorization': 'Bearer ' + hashs.access_token};
  postRequest(url, {}, headers, (res)=>{
    if( res.error ){
      $("#warning").text(res.error).removeClass("hidden").addClass("visible");
      return;
    }
    $("#name").text(res.name)
    $("#thumbnail").attr('src', res.picture);
  },(e)=>{
    console.log(e.responseText);
    $("#warning").text(e.responseText).removeClass("hidden").addClass("visible");
  });
};

var getHashs = function() {
  hashs = parseUrlVars(location.hash);
  return hashs;
};

var parseUrlVars = function(param) {
  if( param.length < 1 ) {
    return {};
  }

  var hash = param;
  if( hash.slice(0, 1) == '#' || hash.slice(0, 1) == '?' ) {
    hash = hash.slice(1);
  }
  var hashs  = hash.split('&');
  var vars = {};
  for( var i = 0 ; i < hashs.length ; i++ ){
    var array = hashs[i].split('=');
    vars[array[0]] = array[1];
  }

  return vars;
}
