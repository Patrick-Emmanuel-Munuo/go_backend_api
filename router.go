// require dependencies
var express = require("express");
var path = require("path");
var { controllers } = require("../controllers");
var router = express.Router();
var {encrypt,decrypt} = require("../helper");

//start backend api router //pass
router.get("/", async(request, responce) => {
  try{
    responce.send(encrypt({success:true,message:"vartrick backend api application router" }));
    }catch(error){
    responce.send(encrypt({success:false, message: error.message}));
  }});

//generate OTP router
router.get("/generate-otp", async(request, responce) => {
  try{
    responce.send(encrypt(controllers.generateOTP()));
  }catch(error){
    responce.send(encrypt({success:false, message: error.message}));
  }
});
//send-email 
router.post("/send-mail", async(request, responce) => {
  try{
    responce.send(encrypt(await controllers.sendMail(decrypt(request.body))));
  }catch(error){
    responce.send(encrypt({success:false, message: error.message}));
  }
});
//send-sms
router.post("/send-sms", async(request, responce) => {
  try{
    responce.send(encrypt(await controllers.sendMessage(decrypt(request.body))));
  }catch(error){
    responce.send(encrypt({success:false, message: error.message}));
  }
});
//send otp-via mail or sms
router.post("/send-otp", async(request, responce) => {
  try{
    responce.send(encrypt(await controllers.sendOTP(decrypt(request.body))));
  }catch(error){
    responce.send(encrypt({success:false, message: error.message}));
  }
});
//file handles upload,download delete and update
router.get("/download", async (request, responce) => {
  try{
    var result = await controllers.file.downloadFile(request.query);
      if(result.success === true){
        responce.download(result.message ,(error) => {
          if(error){
             //responce.send(encrypt({success: false, message:error.message});	
          }else{
             //responce.send(encrypt({success: true, message:"succsessful download file"});
            }
          });
      }else{
        responce.send(encrypt(result));
      }
  }catch(error){
      responce.send(encrypt({success:false, message: error.message}));
  }
});
//upload-file
router.post("/upload-file", async (request, responce) => {
  try{
      responce.send(encrypt(await controllers.file.uploadFile(decrypt(request))));
  }catch(error){
      responce.send(encrypt({success:false, message: error.message}));
  }
});
//upload-files
router.post("/upload-files", async (request, responce) => {
  try{
      responce.send(encrypt(await controllers.file.uploadFiles(decrypt(request.body))));
  }catch(error){
      responce.send(encrypt({success:false, message: error.message}));
  }
});
//delete-file
router.post("/delete-file", async (request, responce) => {
  try{
      responce.send(encrypt(await controllers.file.deleteFile(decrypt(request.body))));
  }catch(error){
      responce.send(encrypt({success:false, message: error.message}));
  }
});
//delete-files
router.post("/delete-files", async (request, responce) => {
  try{
      responce.send(encrypt(await controllers.file.deleteFiles(decrypt(request.body))));
  }catch(error){
      responce.send(encrypt({success:false, message: error.message}));
  }
});
 //exports.router = router;
 module.exports = router;