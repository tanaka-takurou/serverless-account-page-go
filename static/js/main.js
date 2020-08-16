$(document).ready(function() {
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
    window.setTimeout(() => {location.href = "./?page=profile";}, 1000);
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
    SetImgUrl();
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
    url:           App.url
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
function SetImgUrl() {
  if (App.imgUrl.length <= 0) {
    GetImgUrl();
  } else {
    $("#thumbnail").attr("src", App.imgUrl);
  }
}
function GetImgUrl() {
  var token = window.localStorage.getItem("accessToken");
  if (!token) {
    return false;
  }
  const data = {action: 'getimg', token: token};
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    url:           App.url
  })
  .done(function(res) {
    App.imgUrl = res.imgurl;
    if (App.imgUrl.length > 0) {
      $("#thumbnail").attr("src", App.imgUrl);
    } else {
      $("#thumbnail").attr("src", '{{template "default.jpg" .}}');
    }
  })
  .fail(function(e) {
    console.log(e);
  });
}
function OpenModal() {
  $('.large.modal').modal('show');
}
function CloseModal() {
  $('.large.modal').modal('hide');
}
function parseJson (data) {
  var res = {};
  for (i = 0; i < data.length; i++) {
    res[data[i].name] = data[i].value;
  }
  return res;
}
function toBase64 (file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result);
    reader.onerror = error => reject(error);
  });
}
function onConverted () {
  return function(v) {
    App.imgdata = v;
    $('#preview').attr('src', v);
  }
}
function UploadImage(elm) {
  if (!!App.imgdata) {
    $(elm).addClass("disabled");
    putImage();
  } else {
    CloseModal();
  }
}
function putImage() {
  var token = window.localStorage.getItem("accessToken");
  if (!token) {
    return false;
  }
  const file = $('#image').prop('files')[0];
  const data = {action: 'uploadimg', filename: file.name, filedata: App.imgdata, token: token};
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    url:           App.url
  })
  .done(function(res) {
    App.imgUrl = res.imgurl;
    if (App.imgUrl.length > 0) {
      $("#thumbnail").attr("src", App.imgUrl);
    }
  })
  .fail(function(e) {
    console.log(e);
  })
  .always(function() {
    CloseModal();
  });
}
function ChangeImage () {
  const file = $('#image').prop('files')[0];
  toBase64(file).then(onConverted());
}
var App = { imgdata: null, url: location.origin + {{ .ApiPath }}, imgUrl: '' };
