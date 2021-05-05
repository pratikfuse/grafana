import { DataTransformerInfo } from '../../types';
import { map } from 'rxjs/operators';

import { DataTransformerID } from './ids';
import { DataFrame, Field, FieldType } from '../../types/dataFrame';
import { ArrayVector } from '../../vector/ArrayVector';
import { AlignedData, join } from './joinDataFrames';

/* eslint-disable */
export const histogramBucketSizes = [
    .001, .002, .0025, .005,
     .01,  .02,  .025,  .05,
      .1,   .2,   .25,   .5,
       1,    2,     4,    5,
      10,   20,    25,   50,
     100,  200,   250,  500,
    1000, 2000,  2500, 5000,
];
/* eslint-enable */

const histFilter = [null];
const histSort = (a: number, b: number) => a - b;

export interface HistogramTransformerOptions {
  bucketSize?: number; // 0 is auto
}

export const histogramTransformer: DataTransformerInfo<HistogramTransformerOptions> = {
  id: DataTransformerID.histogram,
  name: 'Histogram',
  description: 'Calculate a histogram from input data',
  defaultOptions: {
    fields: {},
  },

  /**
   * Return a modified copy of the series.  If the transform is not or should not
   * be applied, just return the input series
   */
  operator: (options) => (source) =>
    source.pipe(
      map((data) => {
        if (!Array.isArray(data) || data.length === 0) {
          return data;
        }
        return [buildHistogram(data, options)];
      })
    ),
};

const histogramFrameBucketMinFieldName = 'BucketMin';
const histogramFrameBucketMaxFieldName = 'BucketMax';

interface HistogramFields {
  bucketMin: Field;
  bucketMax: Field;
  counts: Field[];
}

/**
 * Given a frame, find the explicit histogram fields
 */
export function getHistogramFields(frame: DataFrame): HistogramFields | undefined {
  let bucketMin: Field | undefined = undefined;
  let bucketMax: Field | undefined = undefined;
  const counts: Field[] = [];
  for (const field of frame.fields) {
    if (field.name === histogramFrameBucketMinFieldName) {
      bucketMin = field;
    } else if (field.name === histogramFrameBucketMaxFieldName) {
      bucketMax = field;
    } else if (field.type === FieldType.number) {
      counts.push(field);
    }
  }
  if (bucketMin && bucketMax && counts.length) {
    return {
      bucketMin,
      bucketMax,
      counts,
    };
  }
  return undefined;
}

export function buildHistogram(frames: DataFrame[], options?: HistogramTransformerOptions): DataFrame {
  let { bucketSize } = options ?? { bucketSize: 0 };

  // if bucket size is auto, try to calc from all numeric fields
  if (!bucketSize) {
    let min = Infinity,
      max = -Infinity;

    // TODO: include field configs!
    for (const frame of frames) {
      for (const field of frame.fields) {
        if (field.type === FieldType.number) {
          for (const value of field.values.toArray()) {
            min = Math.min(min, value);
            max = Math.max(max, value);
          }
        }
      }
    }

    let range = Math.abs(max - min);

    // choose bucket
    for (const size of histogramBucketSizes) {
      if (range / 10 < size) {
        bucketSize = size;
        break;
      }
    }
  }

  const histRound = (v: number) => incrRoundDn(v, bucketSize!);

  let histograms: AlignedData[] = [];

  for (const frame of frames) {
    for (const field of frame.fields) {
      if (field.type === FieldType.number) {
        let fieldHist = histogram(field.values.toArray(), histRound, histFilter, histSort) as AlignedData;
        histograms.push(fieldHist);
      }
    }
  }

  // align histograms
  let joinedHists = join(histograms);

  const bucketMin = new ArrayVector(joinedHists[0]);
  const bucketMax = new ArrayVector(joinedHists[0].map((v) => v + bucketSize!));
  const fields: Field[] = [
    // inclusive
    { name: histogramFrameBucketMinFieldName, values: bucketMin, type: FieldType.number, config: {} },
    // exclusive
    { name: histogramFrameBucketMaxFieldName, values: bucketMax, type: FieldType.number, config: {} },
  ];

  let i = 1;
  for (const frame of frames) {
    for (const field of frame.fields) {
      if (field.type === FieldType.number) {
        fields.push({
          ...field,
          values: new ArrayVector(joinedHists[i]),
        });

        i++;
      }
    }
  }

  return {
    length: bucketMin.length,
    fields,
  };
}

export function incrRound(num: number, incr: number) {
  return Math.round(num / incr) * incr;
}

export function incrRoundUp(num: number, incr: number) {
  return Math.ceil(num / incr) * incr;
}

export function incrRoundDn(num: number, incr: number) {
  return Math.floor(num / incr) * incr;
}

export function histogram(
  vals: number[],
  round: (v: number) => number,
  filter?: any[] | null,
  sort?: ((a: any, b: any) => number) | null
) {
  let hist = new Map();

  for (let i = 0; i < vals.length; i++) {
    let v = vals[i];

    if (v != null) {
      v = round(v);
    }

    let entry = hist.get(v);

    if (entry) {
      entry.count++;
    } else {
      hist.set(v, { value: v, count: 1 });
    }
  }

  filter && filter.forEach((v) => hist.delete(v));

  let bins = [...hist.values()];

  sort && bins.sort((a, b) => sort(a.value, b.value));

  let values = Array(bins.length);
  let counts = Array(bins.length);

  for (let i = 0; i < bins.length; i++) {
    values[i] = bins[i].value;
    counts[i] = bins[i].count;
  }

  return [values, counts];
}
