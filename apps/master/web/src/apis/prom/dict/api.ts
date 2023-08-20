import { Get, Post } from "@/apis/requests";
import type {
  ListDatasourceReply,
  ListDictReply,
  ListDictRequest,
} from "@/apis/prom/dict/dict";

const DictListApi = "/prom/v1/dicts";

const DictDatasourceListApi = "/prom/v1/datasources";

function Datasources() {
  return Get<ListDatasourceReply>(DictDatasourceListApi);
}

function DictList(params: ListDictRequest) {
  return Post<ListDictReply>(DictListApi, params || {});
}

export { Datasources, DictList };
