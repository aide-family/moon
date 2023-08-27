export type IRoute = {
  "zh-CN"?: string;
  "en-US"?: string;
  path?: string;
  icon?: string;
  children?: IRoute[];
};

export const routes: IRoute[] = [
  {
    "zh-CN": "主页",
    "en-US": "Home",
    path: "/",
  },
  {
    "zh-CN": "报警规则",
    "en-US": "Alarm Rules",
    path: "/strategy",
    children: [
      {
        "zh-CN": "套餐",
        "en-US": "Monitor Combo",
        path: "/combo",
      },
      {
        "zh-CN": "规则组",
        "en-US": "Monitor Group",
        path: "/group",
      },
      {
        "zh-CN": "规则",
        "en-US": "Monitor Rule",
        path: "/",
      },
    ],
  },
];
