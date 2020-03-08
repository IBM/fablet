import i18n from '../../i18n';
const echarts = require("echarts/lib/echarts");
require('echarts/lib/chart/bar');
require('echarts/lib/component/tooltip');
require('echarts/lib/component/title');

class BlockEventChart {
    constructor(elementID) {
        this.elementID = elementID;
        this.chart = echarts.init(document.getElementById(this.elementID));

        this.style = {
            grid_lit: { grid: { top: 0, left: 0, right: 0, bottom: 0 } },
            grid_normal: { grid: { top: 50, left: 50, right: 0, bottom: 20} },
            title_lit: { title: { show: false } },
            title_normal: { title: { show: true, text: i18n("block_event_chart_tx"), left: "center" } }
        };
    }

    initChart() {
        this.chart.setOption({
            ...this.style.grid_lit,
            ...this.style.title_lit,
            animation: false,
            xAxis: { type: "time", show: true, name: i18n("block_time"), nameLocation: "end" },
            yAxis: { show: true , name: i18n("block_quantity") },
            series: [
                {
                    type: 'bar',
                    data: []
                }
            ]
        });
    }

    showChart(data) {
        this.chart.setOption({
            series: [{
                data: data
            }]
        });
    }

    resize(fullChart) {
        this.chart.setOption({
            ...(fullChart ? this.style.title_normal : this.style.title_lit),
            ...(fullChart ? this.style.grid_normal : this.style.grid_lit)
        });
        this.chart.resize();
    }
}

export default BlockEventChart;