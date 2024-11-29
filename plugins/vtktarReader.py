import vtk
import tarfile
import lzma

from paraview import util
from vtkmodules.vtkCommonDataModel import vtkStructuredPoints
from vtkmodules.util.vtkAlgorithm import VTKPythonAlgorithmBase
from vtkmodules.numpy_interface import dataset_adapter as dsa
from paraview.util.vtkAlgorithm import smdomain, smhint, smproperty, smproxy


def set_timesteps(algorithm, timesteps):
    executive = algorithm.GetExecutive()
    outInfo = executive.GetOutputInformation(0)

    outInfo.Remove(executive.TIME_STEPS())
    for timestep in timesteps:
        outInfo.Append(executive.TIME_STEPS(), timestep)

    outInfo.Remove(executive.TIME_RANGE())
    outInfo.Append(executive.TIME_RANGE(), timesteps[0])
    outInfo.Append(executive.TIME_RANGE(), timesteps[-1])


def get_timestep(algorithm):
    executive = algorithm.GetExecutive()
    outInfo = executive.GetOutputInformation(0)
    if not outInfo.Has(executive.UPDATE_TIME_STEP()):
        return 0.0
    timestep = outInfo.Get(executive.UPDATE_TIME_STEP())
    return timestep

def parse_bam_header(header):
    # Format:
    # variable ham, level 5, time  1.012500000e+01
    var_str, level_str, time_str = header.split(", ", 2)
    var_name = var_str.split(" ", 2)[1]
    t = float(time_str.split(" ", 1)[1])
    d = {
        "varname": var_name,
        "level": level_str,
        "time": t,
    }
    return d

def get_dimensions(file_content):
    reader = vtk.vtkDataSetReader()
    reader.ReadFromInputStringOn()
    reader.SetBinaryInputString(file_content, len(file_content))
    reader.Update()
    return reader.GetOutput().GetDimensions()


@smproxy.reader(
    name="vtktarReader",
    label="vtktarReader",
    extensions="vtktar",
    file_description="vtktar files",
)
class vtktarReader(VTKPythonAlgorithmBase):
    def __init__(self):
        VTKPythonAlgorithmBase.__init__(
            self,
            nInputPorts=0,
            nOutputPorts=1,
            outputType="vtkStructuredPoints",
        )
        self._timesteps = []


    @smproperty.stringvector(name="FileName")
    @smdomain.filelist()
    @smhint.filechooser(extensions="vtktar", file_description="vtktar files")
    def SetFileName(self, value):
        self._filename = value
        self.Modified()

    @smproperty.doublevector(
        name="TimestepValues",
        information_only="1",
        si_class="vtkSITimeStepsProperty",
    )
    def GetTimestepValues(self):
        timesteps = self._timesteps
        return timesteps.tolist() if timesteps is not None else None


    def RequestInformation(self, request, inInfo, outInfo):
        if self._filename is None:
            return

        whole_extent = [0, 0, 0]

        self.tar = tarfile.open(self._filename)
        members = self.tar.getmembers()
        for member in members:
            # Header
            file_compressed = self.tar.extractfile(member)
            with lzma.LZMAFile(file_compressed) as file:
                file.readline()
                header_str = file.readline().decode("utf-8").rstrip()
                header = parse_bam_header(header_str)
                self._timesteps += [header["time"]]
            file_compressed.close()
            # Whole extent
            file_compressed = self.tar.extractfile(member)
            with lzma.LZMAFile(file_compressed) as file:
                dimensions = get_dimensions(file.read())
                for i in range(3):
                    if dimensions[i] > whole_extent[i]:
                        whole_extent[i] = dimensions[i]
            file_compressed.close()

        util.SetOutputWholeExtent(self, [0, whole_extent[0], 0, whole_extent[1], 0, whole_extent[2]])

        set_timesteps(self, self._timesteps)

        return 1


    def RequestData(self, request, inInfo, outInfo):
        output = dsa.WrapDataObject(vtkStructuredPoints.GetData(outInfo))

        time = get_timestep(self)
        i = self._timesteps.index(time)
        member = self.tar.getmembers()[i]
        file_content = None
        with lzma.LZMAFile(self.tar.extractfile(member)) as file:
            file_content = file.read()

        reader = vtk.vtkDataSetReader()
        reader.ReadFromInputStringOn()
        reader.ReadAllScalarsOn()
        reader.ReadAllVectorsOn()
        reader.ReadAllFieldsOn()
        reader.ReadAllColorScalarsOn()
        reader.ReadAllTensorsOn()
        reader.ReadAllTCoordsOn()
        reader.SetBinaryInputString(file_content, len(file_content))
        reader.Update()
        output.ShallowCopy(reader.GetOutput())

        return 1