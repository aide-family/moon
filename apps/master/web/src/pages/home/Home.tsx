import React, { useState } from "react";
import AlarmSearch from "@/pages/home/child/AlarmSearch";
import type { sizeType } from "@/pages/home/child/AlarmOption";
import AlarmOption from "@/pages/home/child/AlarmOption";
import AlarmTable from "@/pages/home/child/AlarmTable";
import type { AlarmModalProps } from "@/pages/home/child/AlarmModal";
import AlarmModal from "@/pages/home/child/AlarmModal";

export default function Home() {
  const [pageSize, setPageSize] = useState<sizeType>("default");
  const [modalVisible, setModalVisble] = useState<boolean>(false);
  const [alarmModalProps, setAlarmProps] = useState<
    AlarmModalProps | undefined
  >({
    visible: modalVisible,
    setVisible: setModalVisble,
  });

  return (
    <div>
      <AlarmSearch />
      <AlarmOption
        setSize={setPageSize}
        setAlarmModalProps={setAlarmProps}
        setVisible={setModalVisble}
      />
      <AlarmTable size={pageSize} />
      <AlarmModal
        {...alarmModalProps}
        visible={modalVisible}
        setVisible={setModalVisble}
      />
    </div>
  );
}
