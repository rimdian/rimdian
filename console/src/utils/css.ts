import { css } from '@emotion/css'

const xxs = 4
const xs = 8
const s = 12
const m = 16
const l = 24
const xl = 32
const xxl = 64

export const backgroundColorBase = '#F3F6FC'
export const baseBorderRadius = 4
export const colorPrimary = '#4E6CFF'
export const colorSuccess = '#64DD17'
export const borderColorBase = '#d9d9d9'
export const borderColorSecondary = '#f0f0f0'
export const colorLabel = '#5E6C7D'
export const shadowBase = '0 2px 8px rgba(0, 0, 0, 0.15)'

// colors
const slate = '#64748b'
const gray = '#6b7280'
const zinc = '#71717a'
const neutral = '#737373'
const stone = '#78716c'
const red = '#ef4444'
const orange = '#f97316'
const amber = '#f59e0b'
const yellow = '#eab308'
const lime = '#84cc16'
const green = '#22c55e'
const emerald = '#10b981'
const teal = '#14b8a6'
const cyan = '#06b6d4'
const sky = '#0ea5e9'
const blue = '#3b82f6'
const indigo = '#6366f1'
const violet = '#8b5cf6'
const purple = '#a855f7'
const fuchsia = '#d946ef'
const pink = '#ec4899'
const rose = '#f43f5e'

const CSS = {
  // global styles are injected in App.tsx
  // refresh browser to see global styles changes
  GLOBAL: {
    '#root': {
      height: '100vh'
    },
    body: {
      height: '100%',
      backgroundColor: backgroundColorBase,
      fontSize: '14px',
      fontFamily:
        '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", "Helvetica", "Arial", sans-serif',

      // '-webkit-font-smoothing': 'antialiased',
      // '-moz-osx-font-smoothing': 'grayscale',
      WebkitFontSmoothing: 'antialiased',
      MozOsxFontSmoothing: 'grayscale',

      // '-moz-osx-font-smoothing': 'grayscale',
      // '-webkit-font-smoothing': 'antialiased', // decrease text weight
      // '-webkit-text-stroke': '0.45px', // increase text weight
      color: '#414552',
      // '& .ant-table-tbody': {
      //   background: 'red'
      // },
      '& .ant-table-tbody>tr:hover>td, .tr-hover:hover>td': {
        borderBottom: '1px solid rgba(78, 108, 255, 0.4)'
      },
      // '& .tr-hover:hover>td': {
      // backgroundColor: 'rgba(255, 255, 255, 0.5)',
      // },
      '& .ant-alert': {
        borderTop: 'none',
        borderRight: 'none',
        borderLeft: 'none'
      },
      '& .ant-tag': {
        borderLeftWidth: '2px',
        borderTop: 'none',
        borderRight: 'none',
        borderBottom: 'none'
      },
      '& .ant-table th': {
        fontSize: '12px'
      },
      '& .ant-drawer .ant-drawer-close': {
        position: 'fixed',
        marginLeft: '-75px',
        top: '10px',
        backgroundColor: 'rgba(255, 255, 255, 0.4)',
        color: 'white',
        borderRadius: '50%',
        width: '32px',
        height: '32px'
      },
      // '& .ant-drawer.with-bg': {
      //   '& .ant-drawer-content': {
      //     background: backgroundColorBase
      //   },
      //   '& .ant-drawer-header': {
      //     border: 'none',
      //     background: 'none'
      //   }
      // },

      // .ant-drawer.with-tabs {
      //   .ant-drawer-header {
      //     padding: 0;
      //     min-height: 55px;
      //     overflow: hidden;
      //     border: none !important;

      //     .title {
      //       position: absolute;
      //       left: 24px;
      //       top: 16px;
      //     }

      //     .ant-tabs-nav {
      //       height: 55px;
      //     }
      //   }

      //   .ant-drawer-body {
      //     padding: 0;
      //   }
      // }
      '& .no-header .ant-drawer-header': {
        height: 0,
        overflow: 'hidden',
        paddingTop: 0,
        paddingBottom: 0
      }
    },
    table: {
      width: '100%',
      '& td': {
        verticalAlign: 'top'
      }
    },
    'a, a:visited': {
      textDecoration: 'none',
      color: colorPrimary
    },

    // prismjs override
    'pre[class*="language-"]': {
      margin: '0 !important'
    },
    'code[class*="language-"]': {
      textWrap: 'wrap !important'
    }
  },

  AntD: {
    // https://ant.design/theme-editor
    token: {
      colorPrimary: colorPrimary,
      colorSuccess: colorSuccess,
      borderRadius: baseBorderRadius
    },
    components: {
      Table: {
        // reset table header
        colorFillAlter: '#FFF',
        colorFillContent: '#FFF',
        colorFillSecondary: '#FFF',
        colorTextHeading: colorLabel,
        fontWeightStrong: 500
      },
      Cascader: {
        dropdownHeight: 300
      }
    }
  },

  container: css({
    maxWidth: '1200px',
    margin: '0 auto'
  }),

  grid: css({
    display: 'grid',
    gridAutoFlow: 'column',
    gridAutoColumns: '1fr'
  }),

  XXS: '4px',
  XS: '8px',
  S: '12px',
  M: '16px',
  L: '24px',
  XL: '32px',
  XXL: '64px',
  // font sizes
  font_size_xxs: css({ fontSize: '10px' }),
  font_size_xs: css({ fontSize: '12px' }),
  font_size_s: css({ fontSize: '14px' }),
  font_size_m: css({ fontSize: '16px' }),
  font_size_l: css({ fontSize: '18px' }),
  font_size_xl: css({ fontSize: '20px' }),
  font_size_xxl: css({ fontSize: '24px' }),
  // font weights
  font_weight_light: css({ fontWeight: 300 }),
  font_weight_regular: css({ fontWeight: 400 }),
  font_weight_medium: css({ fontWeight: 500 }),
  font_weight_semibold: css({ fontWeight: 600 }),
  font_weight_bold: css({ fontWeight: 700 }),
  slate: slate,
  gray: gray,
  zinc: zinc,
  neutral: neutral,
  stone: stone,
  red: red,
  orange: orange,
  amber: amber,
  yellow: yellow,
  lime: lime,
  green: green,
  emerald: emerald,
  teal: teal,
  cyan: cyan,
  sky: sky,
  blue: blue,
  indigo: indigo,
  violet: violet,
  purple: purple,
  fuchsia: fuchsia,
  pink: pink,
  rose: rose,

  // text colors
  text_slate: css({ color: slate }),
  text_gray: css({ color: gray }),
  text_zinc: css({ color: zinc }),
  text_neutral: css({ color: neutral }),
  text_stone: css({ color: stone }),
  text_red: css({ color: red }),
  text_orange: css({ color: orange }),
  text_amber: css({ color: amber }),
  text_yellow: css({ color: yellow }),
  text_lime: css({ color: lime }),
  text_green: css({ color: green }),
  text_emerald: css({ color: emerald }),
  text_teal: css({ color: teal }),
  text_cyan: css({ color: cyan }),
  text_sky: css({ color: sky }),
  text_blue: css({ color: blue }),
  text_indigo: css({ color: indigo }),
  text_violet: css({ color: violet }),
  text_purple: css({ color: purple }),
  text_fuchsia: css({ color: fuchsia }),
  text_pink: css({ color: pink }),
  text_rose: css({ color: rose }),
  // opacity
  opacity_0: css({ opacity: 0 }),
  opacity_10: css({ opacity: 0.1 }),
  opacity_20: css({ opacity: 0.2 }),
  opacity_30: css({ opacity: 0.3 }),
  opacity_40: css({ opacity: 0.4 }),
  opacity_50: css({ opacity: 0.5 }),
  opacity_60: css({ opacity: 0.6 }),
  opacity_70: css({ opacity: 0.7 }),
  opacity_80: css({ opacity: 0.8 }),
  opacity_90: css({ opacity: 0.9 }),
  // text align
  text_left: css({ textAlign: 'left' }),
  text_center: css({ textAlign: 'center' }),
  text_right: css({ textAlign: 'right' }),
  // pull right / left
  pull_right: css({ float: 'right' }),
  pull_left: css({ float: 'left' }),
  // padding all
  padding_a_xxs: css({ padding: xxs }),
  padding_a_xs: css({ padding: xs }),
  padding_a_s: css({ padding: s }),
  padding_a_m: css({ padding: m }),
  padding_a_l: css({ padding: l }),
  padding_a_xl: css({ padding: xl }),
  padding_a_xxl: css({ padding: xxl }),
  // padding vertical
  padding_v_xxs: css({ padding: `${xxs}px 0` }),
  padding_v_xs: css({ padding: `${xs}px 0` }),
  padding_v_s: css({ padding: `${s}px 0` }),
  padding_v_m: css({ padding: `${m}px 0` }),
  padding_v_l: css({ padding: `${l}px 0` }),
  padding_v_xl: css({ padding: `${xl}px 0` }),
  padding_v_xxl: css({ padding: `${xxl}px 0` }),
  // padding horizontal
  padding_h_xxs: css({ padding: `0 ${xxs}px` }),
  padding_h_xs: css({ padding: `0 ${xs}px` }),
  padding_h_s: css({ padding: `0 ${s}px` }),
  padding_h_m: css({ padding: `0 ${m}px` }),
  padding_h_l: css({ padding: `0 ${l}px` }),
  padding_h_xl: css({ padding: `0 ${xl}px` }),
  padding_h_xxl: css({ padding: `0 ${xxl}px` }),
  // padding top
  padding_t_xxs: css({ paddingTop: xxs }),
  padding_t_xs: css({ paddingTop: xs }),
  padding_t_s: css({ paddingTop: s }),
  padding_t_m: css({ paddingTop: m }),
  padding_t_l: css({ paddingTop: l }),
  padding_t_xl: css({ paddingTop: xl }),
  padding_t_xxl: css({ paddingTop: xxl }),
  // padding bottom
  padding_b_xxs: css({ paddingBottom: xxs }),
  padding_b_xs: css({ paddingBottom: xs }),
  padding_b_s: css({ paddingBottom: s }),
  padding_b_m: css({ paddingBottom: m }),
  padding_b_l: css({ paddingBottom: l }),
  padding_b_xl: css({ paddingBottom: xl }),
  padding_b_xxl: css({ paddingBottom: xxl }),
  // padding left
  padding_l_xxs: css({ paddingLeft: xxs }),
  padding_l_xs: css({ paddingLeft: xs }),
  padding_l_s: css({ paddingLeft: s }),
  padding_l_m: css({ paddingLeft: m }),
  padding_l_l: css({ paddingLeft: l }),
  padding_l_xl: css({ paddingLeft: xl }),
  padding_l_xxl: css({ paddingLeft: xxl }),
  // padding right
  padding_r_xxs: css({ paddingRight: xxs }),
  padding_r_xs: css({ paddingRight: xs }),
  padding_r_s: css({ paddingRight: s }),
  padding_r_m: css({ paddingRight: m }),
  padding_r_l: css({ paddingRight: l }),
  padding_r_xl: css({ paddingRight: xl }),
  padding_r_xxl: css({ paddingRight: xxl }),
  // margin all
  margin_a_xxs: css({ margin: xxs }),
  margin_a_xs: css({ margin: xs }),
  margin_a_s: css({ margin: s }),
  margin_a_m: css({ margin: m }),
  margin_a_l: css({ margin: l }),
  margin_a_xl: css({ margin: xl }),
  margin_a_xxl: css({ margin: xxl }),
  // margin vertical
  margin_v_xxs: css({ margin: `${xxs}px 0` }),
  margin_v_xs: css({ margin: `${xs}px 0` }),
  margin_v_s: css({ margin: `${s}px 0` }),
  margin_v_m: css({ margin: `${m}px 0` }),
  margin_v_l: css({ margin: `${l}px 0` }),
  margin_v_xl: css({ margin: `${xl}px 0` }),
  margin_v_xxl: css({ margin: `${xxl}px 0` }),
  // margin horizontal
  margin_h_xxs: css({ margin: `0 ${xxs}px` }),
  margin_h_xs: css({ margin: `0 ${xs}px` }),
  margin_h_s: css({ margin: `0 ${s}px` }),
  margin_h_m: css({ margin: `0 ${m}px` }),
  margin_h_l: css({ margin: `0 ${l}px` }),
  margin_h_xl: css({ margin: `0 ${xl}px` }),
  margin_h_xxl: css({ margin: `0 ${xxl}px` }),
  // margin top
  margin_t_xxs: css({ marginTop: xxs }),
  margin_t_xs: css({ marginTop: xs }),
  margin_t_s: css({ marginTop: s }),
  margin_t_m: css({ marginTop: m }),
  margin_t_l: css({ marginTop: l }),
  margin_t_xl: css({ marginTop: xl }),
  margin_t_xxl: css({ marginTop: xxl }),
  // margin bottom
  margin_b_xxs: css({ marginBottom: xxs }),
  margin_b_xs: css({ marginBottom: xs }),
  margin_b_s: css({ marginBottom: s }),
  margin_b_m: css({ marginBottom: m }),
  margin_b_l: css({ marginBottom: l }),
  margin_b_xl: css({ marginBottom: xl }),
  margin_b_xxl: css({ marginBottom: xxl }),
  // margin left
  margin_l_xxs: css({ marginLeft: xxs }),
  margin_l_xs: css({ marginLeft: xs }),
  margin_l_s: css({ marginLeft: s }),
  margin_l_m: css({ marginLeft: m }),
  margin_l_l: css({ marginLeft: l }),
  margin_l_xl: css({ marginLeft: xl }),
  margin_l_xxl: css({ marginLeft: xxl }),
  // margin right
  margin_r_xxs: css({ marginRight: xxs }),
  margin_r_xs: css({ marginRight: xs }),
  margin_r_s: css({ marginRight: s }),
  margin_r_m: css({ marginRight: m }),
  margin_r_l: css({ marginRight: l }),
  margin_r_xl: css({ marginRight: xl }),
  margin_r_xxl: css({ marginRight: xxl }),

  // top menu
  top: css({
    display: 'flex',
    padding: `${l}px 0 ${l}px 0`,
    '& h1': {
      position: 'relative',
      fontSize: '20px',
      margin: 0,
      padding: 0,
      color: 'rgba(0, 0, 0, 0.85)',
      fontWeight: 500
    }
  }),

  topSeparator: css({
    marginLeft: 'auto'
  }),

  // left - right bar
  leftRight: css({
    display: 'flex',
    padding: `${l}px 0 ${l}px 0`,
    '& .separator': {
      marginLeft: 'auto'
    }
  }),

  appIcon: css({
    cursor: 'pointer',
    textAlign: 'center',
    width: '30px',
    height: '30px',
    overflow: 'hidden',
    lineHeight: '30px',
    borderRadius: '3px',
    boxShadow: 'rgba(0, 0, 0, 0.15) 1.95px 1.95px 2.6px',
    '& svg': {
      verticalAlign: 'middle'
    },
    '& img': {
      borderRadius: '3px',
      width: '30px',
      height: '30px'
    }
  }),

  borderBottom: {
    solid1: css({ borderBottom: `1px solid ${borderColorBase}` })
  },
  borderRight: {
    solid1: css({ borderRight: `1px solid ${borderColorBase}` })
  },
  borderLeft: {
    solid1: css({ borderLeft: `1px solid ${borderColorBase}` })
  },

  tableTotals: css({
    fontSize: '14px',
    fontWeight: 'bold',
    backgroundColor: 'rgba(0,0,0,0.01)'
  }),

  blockCTA: css({
    position: 'relative',
    borderRadius: baseBorderRadius,
    textAlign: 'center',
    background: 'linear-gradient(to right, rgba(43, 192, 228, 0.5), rgba(234, 236, 198, 0.5))',
    padding: l + 'px'
  }),

  emptyState: {
    container: css({
      backgroundColor: '#fff',
      borderRadius: baseBorderRadius,
      width: 500,
      margin: '100px auto 100px auto',
      textAlign: 'center',
      padding: '100px 50px'
    }),
    icon: css({
      fontSize: '64px',
      color: 'rgba(56, 239, 125, 0.6)',
      marginBottom: '30px'
    })
  },

  cascaderWide: css({
    width: '600px'
    // height: '300px'
  })
}

export default CSS
