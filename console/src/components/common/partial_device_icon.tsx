import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Popover, Space, Tooltip } from 'antd'
import {
  faGamepad,
  faLaptop,
  faMobileAlt,
  faPuzzlePiece,
  faTabletAlt,
  faTshirt,
  faTv,
  faQuestionCircle
} from '@fortawesome/free-solid-svg-icons'
import {
  faAndroid,
  faApple,
  faChrome,
  faEdge,
  faFirefox,
  faInternetExplorer,
  faLinux,
  faSafari,
  faWindows
} from '@fortawesome/free-brands-svg-icons'
import { Device } from 'interfaces'
import Attribute from './partial_attribute'
import CSS from 'utils/css'

export interface PartialDeviceProps {
  device: Device
}

export const PartialDevice = (props: PartialDeviceProps) => {
  return (
    <Popover
      title={
        <>
          {PartialDeviceTypeIcon(props.device.device_type)}&nbsp;&nbsp;
          {props.device.browser} {props.device.browser_version_major}
        </>
      }
      content={
        <>
          {props.device.os && <Attribute label="OS">{props.device.os}</Attribute>}
          {props.device.resolution && (
            <Attribute label="Resolution">{props.device.resolution}</Attribute>
          )}
          <Attribute label="Ad block">{props.device.ad_blocker ? 'yes' : 'no'}</Attribute>
          <Attribute label="Mobile Webview">{props.device.in_webview ? 'yes' : 'no'}</Attribute>
          {props.device.language && <Attribute label="Language">{props.device.language}</Attribute>}
        </>
      }
    >
      <Space className={CSS.opacity_50}>
        {PartialDeviceTypeIcon(props.device.device_type)}
        {props.device.os && PartialDeviceOSIcon(props.device.os)}
        {props.device.browser && PartialDeviceBrowserIcon(props.device.browser)}
      </Space>
    </Popover>
  )
}

export const PartialDeviceTypeIcon = (deviceType: string) => {
  // "desktop", // doesnt exist in the ua-parser-js lib, added when device is unknown
  // "console",
  // "mobile",
  // "tablet",
  // "smarttv",
  // "wearable",
  // "embedded",
  switch (deviceType) {
    case 'desktop':
      return (
        <Tooltip placement="bottom" title="Desktop">
          <FontAwesomeIcon icon={faLaptop} />
        </Tooltip>
      )
    case 'mobile':
      return (
        <Tooltip placement="bottom" title="Mobile phone">
          <FontAwesomeIcon icon={faMobileAlt} />
        </Tooltip>
      )
    case 'tablet':
      return (
        <Tooltip placement="bottom" title="Tablet">
          <FontAwesomeIcon icon={faTabletAlt} />
        </Tooltip>
      )
    case 'smarttv':
      return (
        <Tooltip placement="bottom" title="Smart TV">
          <FontAwesomeIcon icon={faTv} />
        </Tooltip>
      )
    case 'console':
      return (
        <Tooltip placement="bottom" title="Console">
          <FontAwesomeIcon icon={faGamepad} />
        </Tooltip>
      )
    case 'wearable':
      return (
        <Tooltip placement="bottom" title="Wearable">
          <FontAwesomeIcon icon={faTshirt} />
        </Tooltip>
      )
    case 'embedded':
      return (
        <Tooltip placement="bottom" title="Embedded">
          <FontAwesomeIcon icon={faPuzzlePiece} />
        </Tooltip>
      )
    default:
      return (
        <Tooltip placement="bottom" title="Unknown">
          <FontAwesomeIcon icon={faQuestionCircle} />
        </Tooltip>
      )
  }
}

export const PartialDeviceBrowserIcon = (deviceBrowser: string) => {
  /**
   * Possible values :
   * Amaya, Android Browser, Arora, Avant, Baidu, Blazer, Bolt, Camino, Chimera, Chrome,
   * Chromium, Comodo Dragon, Conkeror, Dillo, Dolphin, Doris, Edge, Epiphany, Fennec,
   * Firebird, Firefox, Flock, GoBrowser, iCab, ICE Browser, IceApe, IceCat, IceDragon,
   * Iceweasel, IE [Mobile], Iron, Jasmine, K-Meleon, Konqueror, Kindle, Links,
   * Lunascape, Lynx, Maemo, Maxthon, Midori, Minimo, MIUI Browser, [Mobile] Safari,
   * Mosaic, Mozilla, Netfront, Netscape, NetSurf, Nokia, OmniWeb, Opera [Mini/Mobi/Tablet],
   * Phoenix, Polaris, QQBrowser, RockMelt, Silk, Skyfire, SeaMonkey, SlimBrowser, Swiftfox,
   * Tizen, UCBrowser, Vivaldi, w3m, Yandex
   */
  // first: chrome, edge, safari, firefox
  if (deviceBrowser.indexOf('Chrom') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceBrowser}>
        <FontAwesomeIcon icon={faChrome} />
      </Tooltip>
    )
  }

  if (deviceBrowser.indexOf('Safari') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceBrowser}>
        <FontAwesomeIcon icon={faSafari} />
      </Tooltip>
    )
  }

  if (deviceBrowser.indexOf('Edge') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceBrowser}>
        <FontAwesomeIcon icon={faEdge} />
      </Tooltip>
    )
  }

  if (deviceBrowser.indexOf('Firefox') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceBrowser}>
        <FontAwesomeIcon icon={faFirefox} />
      </Tooltip>
    )
  }

  if (deviceBrowser.indexOf('Opera') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceBrowser}>
        <svg
          aria-hidden="true"
          focusable="false"
          data-prefix="fab"
          data-icon="opera"
          className="svg-inline--fa fa-opera "
          role="img"
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 496 512"
        >
          <path
            fill="currentColor"
            d="M313.9 32.7c-170.2 0-252.6 223.8-147.5 355.1 36.5 45.4 88.6 75.6 147.5 75.6 36.3 0 70.3-11.1 99.4-30.4-43.8 39.2-101.9 63-165.3 63-3.9 0-8 0-11.9-.3C104.6 489.6 0 381.1 0 248 0 111 111 0 248 0h.8c63.1.3 120.7 24.1 164.4 63.1-29-19.4-63.1-30.4-99.3-30.4zm101.8 397.7c-40.9 24.7-90.7 23.6-132-5.8 56.2-20.5 97.7-91.6 97.7-176.6 0-84.7-41.2-155.8-97.4-176.6 41.8-29.2 91.2-30.3 132.9-5 105.9 98.7 105.5 265.7-1.2 364z"
          />
        </svg>
      </Tooltip>
    )
  }

  if (deviceBrowser.indexOf('Android') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceBrowser}>
        <FontAwesomeIcon icon={faAndroid} />
      </Tooltip>
    )
  }

  // second: ie, opera
  if (deviceBrowser.indexOf('IE') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceBrowser}>
        <FontAwesomeIcon icon={faInternetExplorer} />
      </Tooltip>
    )
  }
}

export const PartialDeviceOSIcon = (deviceOS: string) => {
  //  * AIX, Amiga OS, Android, Arch, Bada, BeOS, BlackBerry, CentOS, Chromium OS, Contiki,
  //  * Fedora, Firefox OS, FreeBSD, Debian, DragonFly, Gentoo, GNU, Haiku, Hurd, iOS,
  //  * Joli, Linpus, Linux, Mac OS, Mageia, Mandriva, MeeGo, Minix, Mint, Morph OS, NetBSD,
  //  * Nintendo, OpenBSD, OpenVMS, OS/2, Palm, PCLinuxOS, Plan9, Playstation, QNX, RedHat,
  //  * RIM Tablet OS, RISC OS, Sailfish, Series40, Slackware, Solaris, SUSE, Symbian, Tizen,
  //  * Ubuntu, UNIX, VectorLinux, WebOS, Windows [Phone/Mobile], Zenwalk
  if (deviceOS.indexOf('Windows') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceOS}>
        <FontAwesomeIcon icon={faWindows} />
      </Tooltip>
    )
  }

  if (deviceOS.indexOf('Mac OS') > -1 || deviceOS === 'iOS') {
    return (
      <Tooltip placement="bottom" title={deviceOS}>
        <FontAwesomeIcon icon={faApple} />
      </Tooltip>
    )
  }

  if (deviceOS === 'Android') {
    return (
      <Tooltip placement="bottom" title="Android">
        <FontAwesomeIcon icon={faAndroid} />
      </Tooltip>
    )
  }

  if (deviceOS.indexOf('Chrom') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceOS}>
        <FontAwesomeIcon icon={faChrome} />
      </Tooltip>
    )
  }

  if (deviceOS.indexOf('Firefox') > -1) {
    return (
      <Tooltip placement="bottom" title={deviceOS}>
        <FontAwesomeIcon icon={faFirefox} />
      </Tooltip>
    )
  }

  if (
    [
      'CentOS',
      'Fedora',
      'FreeBSD',
      'Debian',
      'Gentoo',
      'Linux',
      'Mandriva',
      'Mageia',
      'Mint',
      'OpenBSD',
      'RedHat',
      'Slackware',
      'SUSE',
      'Ubuntu',
      'VectorLinux',
      'Zenwalk'
    ].indexOf(deviceOS) > -1
  ) {
    return (
      <Tooltip placement="bottom" title={deviceOS}>
        <FontAwesomeIcon icon={faLinux} />
      </Tooltip>
    )
  }

  return deviceOS
}
