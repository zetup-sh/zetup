import React, {useState, useEffect} from 'react'
import './App.scss'
import './stars.sass'
import './fog.scss'



const App: React.FC = () => {
  const [curCmd, setCurCmd] = useState()

  const base = "https://raw.github.com/zetup-sh/zetup/master/tools"
  const cmds = {
    curl: `sh -c "$(curl -fsSL ${base}/install.sh)"`,
    wget: `sh -c "$(wget ${base}/install.sh -O -)"`,
    powershell: `Set-ExecutionPolicy Bypass -Scope Process -Force; iex ((New-Object System.Net.WebClient).DownloadString('${base}/install.ps1'))`,
  }

  // useEffect(() => {
  //   ;(async () => {
  //     const tagName = await fetch("https://api.github.com/repos/zetup-sh/zetup/releases")
  //       .then(resp => resp.json())
  //       .then(releases => releases[0].tag_name)
  //     setCurTag(tagName)
  //   })()
  // }, [])

  // use effect so it doesn't add the wrong .active class to
  // installation method tab
  useEffect(() => {
    const OS = getOS()
    setCurCmd(OS === "Windows" ? "powershell" : "curl")
  }, [])


  const copyCmd = () => {
    copyTextToClipboard(cmds[curCmd])
  }

  return (
    <div>
    <div id="fog" />
    <div className="stars-container">
    <div id="stars" />
    <div id="stars2" />
    <div id="stars3" />
    </div>
    <h1>Z</h1>
    <h2 className="tagline">Automated Development Setups</h2>
    <div className="install-menu">
    <h3>Installation</h3>
    <div className="tab top">
      {Object.entries(cmds).map(([label]) => (
        <button
          key={label}
          className={"tablinks"+ (curCmd === label ? " active" : "")}
          onClick={()=> setCurCmd(label)}>
        {label}
       </button>
      ))}
      <div style={{
        display: "flex",
        flex: "1",
      }} />
    </div>
    <div className="tabcontent">
      <div style={{
        display: "flex",
        width: "100%",
      }}>
      <div style={{
        flex: 10,
        overflow: "hidden",
        textOverflow: "ellipsis",
      }}>
      {!cmds[curCmd] ? "Retrieving..." :
      <code style={{margin: "0 auto"}} id="install-cmd">{cmds[curCmd]}</code>
      }
      </div>
      <div style={{
        display: "flex",
        flex: 1,
      }} />
      <div style={{
        display: "flex",
        flex: 1,
      }}>
      <button className="btn" onClick={copyCmd}><i className="fa fa-copy"></i></button>
      </div>
      </div>
    </div>
    </div>
    </div>
  );
}

export default App;

function getOS() {
  var userAgent = window.navigator.userAgent,
      platform = window.navigator.platform,
      macosPlatforms = ['Macintosh', 'MacIntel', 'MacPPC', 'Mac68K'],
      windowsPlatforms = ['Win32', 'Win64', 'Windows', 'WinCE'],
      iosPlatforms = ['iPhone', 'iPad', 'iPod']
  let os = ""

  if (macosPlatforms.indexOf(platform) !== -1) {
    os = 'Mac OS';
  } else if (iosPlatforms.indexOf(platform) !== -1) {
    os = 'iOS';
  } else if (windowsPlatforms.indexOf(platform) !== -1) {
    os = 'Windows';
  } else if (/Android/.test(userAgent)) {
    os = 'Android';
  } else if (!os && /Linux/.test(platform)) {
    os = 'Linux';
  }

  return os;
}


// https://stackoverflow.com/a/30810322
function fallbackCopyTextToClipboard(text) {
  var textArea = document.createElement("textarea");
  textArea.value = text;
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();

  try {
    var successful = document.execCommand('copy');
    var msg = successful ? 'successful' : 'unsuccessful';
    console.log('Fallback: Copying text command was ' + msg);
  } catch (err) {
    console.error('Fallback: Oops, unable to copy', err);
  }

  document.body.removeChild(textArea);
}

function copyTextToClipboard(text) {
  if (!navigator.clipboard) {
    fallbackCopyTextToClipboard(text);
    return;
  }
  navigator.clipboard.writeText(text).then(function() {
    console.log('Async: Copying to clipboard was successful!');
  }, function(err) {
    console.error('Async: Could not copy text: ', err);
  });
}