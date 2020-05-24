$(document).ready(function() {
  var
    $headers     = $('body > div > div > h2'),
    $header      = $headers.first(),
    ignoreScroll = false,
    timer;

  $(window)
    .on('resize', function() {
      clearTimeout(timer);
      $headers.visibility('disable callbacks');

      $(document).scrollTop( $header.offset().top );

      timer = setTimeout(function() {
        $headers.visibility('enable callbacks');
      }, 500);
    });
  $headers
    .visibility({
      once: false,
      checkOnRefresh: true,
      onTopPassed: function() {
        $header = $(this);
      },
      onTopPassedReverse: function() {
        $header = $(this);
      }
    });
  let params = new URLSearchParams(document.location.search.substring(1));
  let page = params.get("page");
  if (page == "profile") {
    getuser();
  } else if (page == "logout") {
    logout();
  }
});

var login = function() {
  $("#submit").addClass('disabled');
  var name = $("#nickName").val();
  var pass = $('#password').val();
  if (!name | !pass) {
    return false;
  }
  const action = "login";
  const data = {action, name, pass};
  request(data, (res)=>{
    window.localStorage.setItem("accessToken", res.token);
    window.setTimeout(() => {location.href = "/?page=profile";}, 1000);
  }, onError);
};

var changepass = function() {
  var token = window.localStorage.getItem("accessToken");
  if (!token) {
    return false;
  }
  if (!checkPass($("#newpassword").val())) {
    $("#newpassword").val('');
    $("#passwarning").removeClass("hidden").addClass("visible");
    return
  }
  $("#submit").addClass('disabled');
  var pass = $('#password').val();
  var newpass = $('#newpassword').val();
  if (!token | !pass | !newpass) {
    return false;
  }
  const action = "changepass";
  const data = {action, token, pass, newpass};
  request(data, (res)=>{
    console.log(res);
    $("#info").removeClass("hidden").addClass("visible");
  }, onError);
};

var signup = function() {
  if (!checkPass($("#password").val())) {
    $("#password").val('');
    $("#passwarning").removeClass("hidden").addClass("visible");
    return
  }
  $("#submit").addClass('disabled');
  var mail = $("#email").val();
  var name = $("#nickName").val();
  var pass = $("#password").val();
  if (!mail | !name | !pass) {
    return false;
  }
  const action = "signup";
  const data = {action, mail, name, pass};
  request(data, (res)=>{
    console.log(res);
    $("#info").removeClass("hidden").addClass("visible");
  }, onError);
};

var activate = function() {
  $("#submit").addClass('disabled');
  var name = $("#nickName").val();
  var code = $("#activationKey").val();
  if (!name | !code) {
      return false;
  }
  const action = "confirmsignup";
  const data = {action, name, code};
  request(data, (res)=>{
    console.log(res);
    $("#info").removeClass("hidden").addClass("visible");
  }, onError);
};

var getuser = function() {
  var token = window.localStorage.getItem("accessToken");
  if (!token) {
    $("#settings").addClass("hidden");
    $("#warning").text("You are not logged in yet.").removeClass("hidden").addClass("visible");
    return false;
  }
  const action = "getuser";
  const data = {action, token};
  request(data, (res)=>{
    console.log(res);
    $("#name").text(res.name);
  }, onError);
};

var logout = function() {
  var token = window.localStorage.getItem("accessToken");
  if (!token) {
    return false;
  }
  const action = "logout";
  const data = {action, token};
  request(data, (res)=>{
    console.log(res);
    window.localStorage.setItem("accessToken", "");
    console.log("return to top");
  }, onError);
};

var request = function(data, callback, onerror) {
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    url:           {{ .Api }}
  })
  .done(function(res) {
    callback(res);
  })
  .fail(function(e) {
    onerror(e);
  });
};

var checkPass = function(s) {
  const ra = new RegExp(/[0-9]+/);
  const rb = new RegExp(/[a-z]+/);
  const rc = new RegExp(/[A-Z]+/);
  const rd = new RegExp(/[#\\(\\)_\\-\\@\\%\\#\\&\\$\\^\\*]+/);
  return s.length > 7 && ra.test(s) && rb.test(s) && rc.test(s) && rd.test(s)
};

var onError = function(e) {
  console.log(e.responseJSON.message);
  $("#warning").text(e.responseJSON.message).removeClass("hidden").addClass("visible");
  $("#submit").removeClass('disabled');
};
