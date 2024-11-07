const signalingServerURL = "https://autodiscovery-signaling.app-builder-on-prem.net"

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
  
  pc.addEventListener('connectionstatechange', event => {
    if (pc.connectionState === 'connected') {
      console.log("CONNECTED")
    }
  })
  
  let localChannel;
  
  pc.addEventListener("datachannel", event => {
    event.channel.addEventListener("message", event => {
      console.log("got message", event.data)
    })
  })
  
  let c = pc.createDataChannel("send")
  c.addEventListener("open", event => {
    localChannel = event.currentTarget;
    c.send("hi")
  })
  
  let pollAnswerHandle;
  let pollAnswerCandidatesHandle;
  
  let sessionInput = document.getElementById("session");
  
  let sendBtn = document.getElementById("send-btn");
  sendBtn.onclick = event => {
    localChannel.send("hi")
  }
  
  let clientBtn = document.getElementById("client-btn");
  clientBtn.onclick = async () => {
    const offerDescription = await pc.createOffer();
    await pc.setLocalDescription(offerDescription);
  
    const offer = {
      sdp: offerDescription.sdp,
      type: offerDescription.type
    };
  
    await createSession(offer);

    pc.addEventListener('icecandidate', async event => {
      if (event.candidate) {
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
      await pc.setRemoteDescription(answerDesc);
    }
  };
  
  let pollAnswerCandidates = async () => {
    let session = await getSession();
    if (session && session.answer_candidates.length > 0) {
      clearInterval(pollAnswerCandidatesHandle);
  
      for (const c of session.answer_candidates) {
        const candidate = new RTCIceCandidate(c);
        await pc.addIceCandidate(candidate);
      }
    }
  };
  
  let getSession = async () => {
    return (await fetch(`${signalingServerURL}/getSession?id=` + sessionInput.value)).json()
  }
  