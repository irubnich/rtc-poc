const signalingServerURL = "https://autodiscovery-signaling.app-builder-on-prem.net"
//const signalingServerURL = "http://localhost:8080"

const config = {
    iceServers: [
      {
        urls: ['stun:stun1.l.google.com:19302', 'stun:stun2.l.google.com:19302'],
      },
    ],
    iceCandidatePoolSize: 10,
  }
  
  const pc = new RTCPeerConnection(config);
  pc.addTransceiver("audio")
  
  let logsBlock = document.getElementById("logs")
  let log = msg => {
    logsBlock.innerHTML += `${msg}<br>` 
  }

  pc.addEventListener('connectionstatechange', event => {
    log(`new connection state: ${pc.connectionState}`) 
  })
  
  let localChannel;
  
  pc.addEventListener("datachannel", event => {
    event.channel.addEventListener("message", event => {
      log(`got message on datachannel: ${event.data}`)
    })
  })
  
  let c = pc.createDataChannel("send")
  c.addEventListener("open", event => {
    log(`data channel opened`)

    localChannel = event.currentTarget;
    c.send("hello from browser")
  })
  
  let pollAnswerHandle;
  let pollAnswerCandidatesHandle;
  
  let sessionInput = document.getElementById("session");
  
  let sendBtn = document.getElementById("send-btn");
  sendBtn.onclick = event => {
    log(`sending a message over the data channel`)
    localChannel.send("hi")
  }
  
  let clientBtn = document.getElementById("client-btn");
  clientBtn.onclick = async () => {
    log(`creating offer`)
    const offerDescription = await pc.createOffer();
    await pc.setLocalDescription(offerDescription);
  
    const offer = {
      sdp: offerDescription.sdp,
      type: offerDescription.type
    };
  
    log(`creating session: ${sessionInput.value}`)
    await createSession(offer);

    pc.addEventListener('icecandidate', async event => {
      if (event.candidate) {
        log(`got ICE candidate: ${event.candidate.toJSON().candidate}`)
        await addOfferCandidate(event.candidate);
      }
    })
  
    pollAnswerHandle = setInterval(pollAnswer, 2000);
    pollAnswerCandidatesHandle = setInterval(pollAnswerCandidates, 2000);
  };
  
  var createSession = async (session) => {
    await fetch(`${signalingServerURL}/createSession?id=${sessionInput.value}`, {
      method: "POST",
      body: JSON.stringify(session)
    })
  };
  
  var addOfferCandidate = async (candidate) => {
    await fetch(`${signalingServerURL}/addOfferCandidate?id=` + sessionInput.value, {
      method: "POST",
      body: JSON.stringify(candidate)
    })
  };
  
  let pollAnswer = async () => {
    let session = await getSession();
    if (session && session.answer) {
      clearInterval(pollAnswerHandle);

      const answerDesc = new RTCSessionDescription(session.answer);
      log(`got answer!`)
      await pc.setRemoteDescription(answerDesc);
    }
  };
  
  let pollAnswerCandidates = async () => {
    let session = await getSession();
    if (session && session.answer_candidates.length > 0) {
      clearInterval(pollAnswerCandidatesHandle);
  
      for (const c of session.answer_candidates) {
        log(`got answer candidate!`)
        const candidate = new RTCIceCandidate(c);
        await pc.addIceCandidate(candidate);
      }
    }
  };
  
  let getSession = async () => {
    return (await fetch(`${signalingServerURL}/getSession?id=` + sessionInput.value)).json()
  }
  