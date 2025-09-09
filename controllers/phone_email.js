const path = require('path');
const moment = require("moment")
const configurations = require("../configuration/config");
const { SerialPort } = require("serialport");
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));
const moderm_data ={
      path: 'COM5',
      manufacturer: 'ZTE Corporation',
      serialNumber: '6&3345B05&2&0000',
      pnpId: 'USB\\VID_19D2&PID_0016&MI_00\\6&3345B05&2&0000',
      locationId: '0000.0014.0000.001.000.000.000.000.000',
      friendlyName: 'ZTE Diagnostics Interface (COM6)',
      vendorId: '19D2',
      productId: '0016'
}

/* create and export asynchronous function to send email */
 async function sendMail(options) {
   try {
    var message, to, subject, html, attachments;
    if(options){
      text = options.message;
      to = options.to;
      subject = (!options.subject?text:options.subject);
      html = (!options.html?`<h2>${text}</h2>`:options.html);
      attachments = options.attachments || [];
      if(!text || !to){
          return { success: false, message: "both mail to and message required can't be empty" };
      }else{
        	/*if(options.filePath){
                var filepath = path.join(__dirname, '../../../../public/' + options.filePath );
        		mail['attachments'] = [{ path : filepath }]
        	}*/
        	//console.log(mail)
        let result = await configurations.mailTransporter.sendMail({
          from : configurations.mailSender,
          to,
          subject,
          text,
          html,
          attachments: attachments.length > 0 ? attachments : undefined
        })
        //console.log({ success: true, message: result })
        return { success: true, message: result };
      }     
    }else{
      return { success: false, message: "function options parameter required can't be empty" };
    }
    }catch (error) {
        return { success: false, message: error.message };
    }
}
/* create and export asynchronous function to send message */
async function sendMessage(options) {
   try {
    var message, to;
      if(options){
        message = options.message, to = options.to;
        if(!message || !to){
            return { success: false, message: "both to and message required can't be empty" };
        }else {
            if((message === '') && (to.length > 0)){
                return { success: false, message: "message can't be empty" };
            }else if(!(message === '') && !(to.length > 0)){
                return { success: false, message: "message and to can't be empty" }; 
            }else if(!Array.isArray(to)){
                return { success: false, message: "to must be an array" };   
            }else{
                //Send Message
                 let result = await configurations.senderSMS.send({
                    to ,
                    message ,
                    from: configurations.senderId
                });
                var results = {
                  created_date: moment().format("YYYY-MM-DDTHH:mm:ssZ"),
                  created_by:"system",
                  updated_by:null,
                  updated_date: moment().format("YYYY-MM-DDTHH:mm:ssZ"),
                  schedule_time:"null",
                  sender_id:configurations.senderId,
                  text:message,
                  summary: result.SMSMessageData.Message,
                  totalCost:
                    result.SMSMessageData.Recipients.reduce(
                      (total, recipient) => total + parseFloat(recipient.cost.replace("TZS ", "")),
                      0
                    ).toFixed(2) + " TZS",
                  recipients: result.SMSMessageData.Recipients.map((recipient) => ({
                    phoneNumber: recipient.number,
                    messageId: recipient.messageId,
                    cost: recipient.cost,
                    status: recipient.status,
                    statusCode: recipient.statusCode,
                    messageParts: recipient.messageParts,
                  })),
                }
                var database = ""
                return { success: true,message:results}; 
                
            }
        }
    }else{
        return { success: false, message: "function options parameter required can't be empty" };
    }
    }catch (error) {
        return { success: false, message: error.message };
    }
}

//
const openPort = async (port) => {
  try {
    await new Promise((resolve, reject) => {
      port.open((err) => {
        if (err) {
          console.log({ success: false, message: "Failed to open port: " + err.message });
          reject(err);
        } else {
          console.log({ success: true, message: "Port opened successfully: " + port.path });
          resolve();
        }
      });
    });
  } catch (error) {
    console.log({ success: false, message: "Error in openPort: " + error.message });
  }
};

/**Read SMS messages from modem (COM6 if available) */
async function readMessage() {
  try {
    const ports = await SerialPort.list();
    const comPort = ports.find(
      p =>
        p.manufacturer?.includes("ZTE") &&
        p.path.toUpperCase() === moderm_data.path.toUpperCase()
    );

    if (!comPort) {
      return { success: false, message: `${moderm_data.path} not found. Please connect modem.` };
    }

    const portModem = new SerialPort({
      path: comPort.path,
      baudRate: 9600,
      parity: "none",
      stopBits: 1,
      dataBits: 8,
      autoOpen: false,
    });

    await openPort(portModem);
    console.log({ success: true, message: `Modem port opened: ${comPort.path}` });

    // Initialize modem
    const initCommands = ["AT\r", "AT+CMGF=1\r", "AT+CNMI=1,2,0,0,0\r"];
    for (const cmd of initCommands) {
      portModem.write(cmd);
      await sleep(200);
    }

    let bufferData = "";
    portModem.on("data", async (buffer) => {
      bufferData += buffer.toString();
      if (bufferData.includes("\r\n")) {
        const lines = bufferData.split("\r\n").filter(l => l.trim());
        bufferData = ""; // reset
        for (let i = 0; i < lines.length; i++) {
          //console.log("data: ", lines[i]);
          if (lines[i].startsWith("+CMT:")) {
            const header = lines[i];
            const sms_text = lines[i + 1] || ""; // SMS text is next line
            const parts = header.split(",");
            const receiver = parts[0].split(":")[1].replace(/"/g, "").trim();
            // Date and time
            // Format date to YYYY-MM-DD
            //const formatted_date = moment(receiver_date, "YY/MM/DD").format("DD-MM-YYYY");
            const receiver_date = moment(parts[2]?.replace(/"/g, "").trim().split(",")[0], "YY/MM/DD")?.format("DD-MM-YYYY");
            const receiver_time = parts[3]?.replace(/"/g, "").trim().split(",")[0]?.split("+")[0]; 

            const saved_data = {
              receiver,
              receiver_date,
              receiver_time,
              receiver_text: String(sms_text ?? "").trim(),  // safely handle undefined or null
              created_date: moment().format("YYYY-MM-DDTHH:mm:ssZ"),
              created_by: "system",
              status: "received",
              method: "local_via_modem"
            };
            console.log({ success: true, message: saved_data });
            // Optional: auto-reply
            try {
              const responce = await sendMessage_local({ to: [receiver], message: "thanks for auto-reply"  });
              if(responce.success){
                console.log({
                  success: true,
                  status: 'Auto-replied',
                  message: responce.message
                });
              }else{
                console.log({
                  success: false,
                  status: 'Auto-replied failed',
                  message: responce.message
                });
              }
            } catch (err) {
              console.log({ success: false, message: "Failed to send auto-reply: " + err.message });
            }
          }
        }
      }
    });

    portModem.on("error", (err) => console.error("Modem error:", err.message));
    return { success: true, message: `Listening on ${comPort.path}...` };
  } catch (error) {
    console.log({ success: false, message: error.message });
    return { success: false, message: error.message };
  }
}

/**Send SMS via local modem (COM8 if available) */
async function sendMessage_local(options) {
  try {
    const phones = Array.isArray(options.to) ? options.to : [options.to || "0760449295"];
    if (phones.length === 0) {
      return { success: false, message: "No phone numbers provided." };
    }
    const message = options.message
      ? `${options.message}\rsent at ${moment().format("YYYY-MM-DDTHH:mm:ssZ")}`
      : `Message sent at ${moment().format("YYYY-MM-DDTHH:mm:ssZ")}`;

    const ports = await SerialPort.list();
    const comPort = ports.find(
      p => p.manufacturer?.includes("ZTE") && p.path.toUpperCase() === moderm_data.path.toUpperCase()
    );
    if (!comPort) return { success: false, message: `${moderm_data.path} not found. Please connect modem.` };

    const portModem = new SerialPort({ 
        path: comPort.path,
        baudRate: 9600,
        parity: "none",
        stopBits: 1,
        dataBits: 8,
        autoOpen: false 
      });
    await openPort(portModem);
    // Initialize modem
    portModem.write("AT\r");
    await sleep(200);
    portModem.write("AT+CMGF=1\r"); // Text mode
    await sleep(200);
    const results = [];
    // Helper function to send a single SMS and wait for modem response
    const sendSingleSMS = (phone, message) => new Promise(async (resolve) => {
      let bufferData = "";
      const onData = (data) => {
        bufferData += data.toString();
        if (bufferData.includes("+CMGS:")) {
          portModem.off("data", onData);
          resolve({ phone, status: "sent", response: bufferData.trim() });
        } else if (bufferData.includes("+CMS ERROR:")) {
          portModem.off("data", onData);
          const errorCode = bufferData.match(/\+CMS ERROR: (\d+)/)?.[1] || "unknown";
          resolve({ phone, status: "failed", errorCode, response: bufferData.trim() });
        }
      };
      portModem.on("data", onData);
      try {
        portModem.write(`AT+CMGS="${phone}"\r`);
        await sleep(200);
        portModem.write(message + "\r");
        await sleep(200);
        portModem.write(Buffer.from([0x1A])); // Ctrl+Z
      } catch (err) {
        portModem.off("data", onData);
        resolve({ phone, status: "failed", error: err.message });
      }
      // Timeout in case modem doesn’t respond
      setTimeout(() => {
        portModem.off("data", onData);
        resolve({ phone, status: "unknown", response: bufferData.trim() });
      }, 3000);
    });

    // Send SMS to all phones
    for (const rawPhone of phones) {
      const phone = rawPhone.startsWith("+") ? rawPhone : "+255" + rawPhone.replace(/^0/, "");
      const result = await sendSingleSMS(phone, message);
      result.sent_date = moment().format("YYYY-MM-DDTHH:mm:ssZ");
      results.push(result);
    }

    const saved_data = {
      sender_id: "local",
      created_by: "system",
      updated_date: moment().format("YYYY-MM-DDTHH:mm:ssZ"),
      method: "local_via_modem",
      cost: "0.00",
      messages: results
    };
    return { success: true, message: saved_data };
  } catch (error) {
    console.log({ success: false, message: error.message });
    return { success: false, message: error.message };
  }
}

// Start reading messages
readMessage();
//sendMessage_local({ to: "0760449295", message: "Test message from local modem" });

module.exports = {
  sendMail,
  sendMessage,
  sendMessage_local
};
